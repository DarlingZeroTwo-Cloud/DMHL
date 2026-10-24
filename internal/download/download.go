package download

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Progress 下载进度回调
type Progress struct {
	Total      int64 `json:"total"`
	Done       int64 `json:"done"`
	Failed     int64 `json:"failed"`
	Skipped    int64 `json:"skipped"`
	CurrentURL string `json:"currentUrl"`
}

// Manager 并发下载器
type Manager struct {
	Concurrency int
	client      *http.Client
}

func New() *Manager {
	return &Manager{
		Concurrency: 16,
		client:      &http.Client{Timeout: 120 * time.Second},
	}
}

// Item 一个待下载文件
type Item struct {
	URL     string
	Dest    string
	SHA1    string // 空=不校验
	Size    int64
}

// Do 并发下载全部 item，返回失败列表
func (m *Manager) Do(items []Item, onProgress func(*Progress)) []error {
	var total int64
	for _, it := range items {
		total += it.Size
	}
	prog := &Progress{Total: total}

	sem := make(chan struct{}, m.Concurrency)
	var wg sync.WaitGroup
	var errsMu sync.Mutex
	var errs []error

	for _, it := range items {
		wg.Add(1)
		go func(it Item) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 已存在且 hash 对，跳过
			if st, err := os.Stat(it.Dest); err == nil && st.Size() == it.Size && it.SHA1 == "" {
				atomic.AddInt64(&prog.Skipped, 1)
				return
			}
			if it.SHA1 != "" && fileSHA1(it.Dest) == it.SHA1 {
				atomic.AddInt64(&prog.Skipped, 1)
				return
			}

			if err := m.downloadOne(it); err != nil {
				errsMu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", it.URL, err))
				errsMu.Unlock()
				atomic.AddInt64(&prog.Failed, 1)
				return
			}
			atomic.AddInt64(&prog.Done, it.Size)
			if onProgress != nil {
				onProgress(prog)
			}
		}(it)
	}
	wg.Wait()
	return errs
}

func (m *Manager) downloadOne(it Item) error {
	_ = os.MkdirAll(filepath.Dir(it.Dest), 0755)
	tmp := it.Dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	resp, err := m.client.Get(it.URL)
	if err != nil {
		f.Close()
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		f.Close()
		return errors.New("HTTP " + it.URL)
	}
	n, err := io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		return err
	}
	if it.Size > 0 && n != it.Size {
		return fmt.Errorf("size mismatch: got %d want %d", n, it.Size)
	}
	if it.SHA1 != "" {
		got := fileSHA1(tmp)
		if got != it.SHA1 {
			os.Remove(tmp)
			return fmt.Errorf("sha1 mismatch: want %s got %s", it.SHA1, got)
		}
	}
	return os.Rename(tmp, it.Dest)
}

func fileSHA1(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha1.New()
	_, _ = io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil))
}

// GetBytes 下载小文件到内存
func (m *Manager) GetBytes(url string) ([]byte, error) {
	resp, err := m.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("HTTP " + url)
	}
	return io.ReadAll(resp.Body)
}
