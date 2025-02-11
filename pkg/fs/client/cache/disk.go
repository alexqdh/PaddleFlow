/*
Copyright (c) 2021 PaddlePaddle Authors. All Rights Reserve.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"paddleflow/pkg/fs/utils/mount"
)

const CacheDir = "datacache"

type cacheItem struct {
	size    int64
	expTime time.Time
}

type diskCache struct {
	sync.RWMutex
	dir      string
	capacity int64
	used     int64
	expire   time.Duration
	keys     map[string]*cacheItem
}

type DiskConfig struct {
	Dir    string
	Mode   os.FileMode
	Expire time.Duration
}

func NewDiskCache(config *DiskConfig) *diskCache {
	if config == nil || config.Dir == "" || config.Dir == "/" {
		return nil
	}

	if config != nil {
		d := &diskCache{
			dir:    config.Dir,
			keys:   make(map[string]*cacheItem),
			expire: config.Expire,
		}
		// TODO: 报错往上抛
		os.MkdirAll(config.Dir, 0755)
		d.updateCapacity()
		// 后续加stop channel
		go func() {
			for {
				d.clean()
				time.Sleep(10 * time.Second)
			}
		}()

		return d
	}
	return nil
}

func (c *diskCache) load(key string) (ReadCloser, bool) {
	if c.dir == "" {
		return nil, false
	}

	if !c.exist(key) {
		return nil, false
	}

	path := c.cachePath(key)
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return nil, false
	}

	if err != nil {
		log.Errorf("stat cache file[%s] failed: %v", path, err)
		return nil, false
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	return f, true
}

func (c *diskCache) save(key string, buf []byte) {
	if c.dir == "" {
		return
	}
	cacheSize := int64(len(buf))
	if c.used+cacheSize >= c.capacity {
		// todo：clean支持带参数，释放多少容量。
		c.clean()
	}

	// 清理之后还是没有足够的容量，则跳过
	if c.used+cacheSize >= c.capacity {
		return
	}
	path := c.cachePath(key)
	c.createDir(filepath.Dir(path))
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		log.Errorf("open tmp file[%s] failed: %v", tmp, err)
		return
	}
	_, err = f.Write(buf)
	if err != nil {
		log.Errorf("write tmp file[%s] failed: %v", tmp, err)
		_ = f.Close()
		_ = os.Remove(tmp)
		return
	}
	err = f.Close()
	if err != nil {
		log.Errorf("close tmp file[%s] failed: %v", tmp, err)
		_ = os.Remove(tmp)
		return
	}
	err = os.Rename(tmp, path)
	if err != nil {
		log.Errorf("rename file %s -> %s failed: %v", tmp, path, err)
		_ = os.Remove(tmp)
		return
	}

	c.Lock()
	c.keys[key] = &cacheItem{
		expTime: time.Now().Add(c.expire),
		size:    cacheSize,
	}
	c.Unlock()
	log.Debugf("diskCache save[%s] succeed", key)
	return
}

func (c *diskCache) delete(key string) {
	path := c.cachePath(key)
	c.Lock()
	if c.keys[key] != nil {
		delete(c.keys, key)
	}
	c.Unlock()
	if path != "" {
		go os.Remove(path)
	}
}

func (c *diskCache) clean() {
	// 1. 首先清理掉已过期文件
	for key, cache := range c.keys {
		if time.Now().Sub(cache.expTime) >= 0 {
			c.delete(key)
		}
	}

	log.Debugf("the c.dir is [%s]", c.dir)
	if c.dir == "/" || c.dir == "" {
		return
	}
	// 2. 清理目录下存在，但是c.keys中不存在的的文件,
	filepath.Walk(filepath.Join(c.dir, CacheDir), func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		key := c.getKeyFromCachePath(path)
		log.Debugf("clean dis cache key is %s and path is %s", key, path)
		if c.keys[key] != nil {
			return nil
		}
		err = os.Remove(path)
		return err
	})
	c.updateCapacity()
}

func (c *diskCache) cachePath(key string) string {
	return filepath.Join(c.dir, CacheDir, key)
}

func (c *diskCache) getKeyFromCachePath(path string) string {
	return strings.TrimPrefix(path, filepath.Join(c.dir, CacheDir)+"/")
}

func (c *diskCache) createDir(dir string) {
	os.MkdirAll(dir, 0755)
}

func (c *diskCache) exist(key string) bool {
	c.RLock()
	defer c.RUnlock()
	if value, ok := c.keys[key]; !ok {
		return false
	} else if value.expTime.Sub(time.Now()) <= 0 {
		log.Debugf("expire key %s", key)
		return false
	}
	return true
}

func (c *diskCache) updateCapacity() error {
	output, err := mount.ExecCmdWithTimeout("df", []string{"-k", c.dir})
	if err != nil {
		log.Errorf("df %s %v ", c.dir, err)
		return err
	}

	buf := bufio.NewReader(bytes.NewReader(output))
	for {
		if line, _, err := buf.ReadLine(); err != nil {
			return err
		} else {
			if strings.HasPrefix(string(line), "Filesystem") {
				continue
			}
			strSlice := strings.Fields(strings.TrimSpace(string(line)))
			if len(strSlice) < 6 {
				continue
			}

			total, err := strconv.ParseInt(strSlice[1], 10, 64)
			if err != nil {
				log.Errorf("parse str[%s] failed: %v", strSlice[1], err)
				return err
			}
			used, err := strconv.ParseInt(strSlice[2], 10, 64)
			if err != nil {
				log.Errorf("parse str[%s] failed: %v", strSlice[2], err)
				return err
			}
			c.Lock()
			c.capacity = total * 1024
			c.used = used * 1024
			c.Unlock()
			return nil
		}
	}
}

var _ Cache = &diskCache{}
