package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"dmhl/internal/auth"
	"dmhl/internal/config"
	"dmhl/internal/instance"
	"dmhl/internal/java"
	"dmhl/internal/mirror"
	"dmhl/internal/mojang"
	"dmhl/internal/skin"
)

// Server HTTP API 服务
type Server struct {
	Cfg       *config.Config
	InstMgr   *instance.Manager
	AuthStore *auth.Store
	SkinMgr   *skin.Manager
}

func New(cfg *config.Config) *Server {
	return &Server{
		Cfg:       cfg,
		InstMgr:   instance.New(cfg.DataDir),
		AuthStore: auth.NewStore(filepath.Join(cfg.DataDir, "accounts.json")),
		SkinMgr:   skin.New(cfg.DataDir),
	}
}

// Routes 注册路由
func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/versions", s.handleVersions)
	mux.HandleFunc("/api/version/detail", s.handleVersionDetail)
	mux.HandleFunc("/api/instances", s.handleInstances)
	mux.HandleFunc("/api/instances/create", s.handleCreateInstance)
	mux.HandleFunc("/api/instances/launch", s.handleLaunch)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/auth/offline", s.handleOfflineLogin)
	mux.HandleFunc("/api/accounts", s.handleAccounts)
	mux.HandleFunc("/api/skin/import", s.handleSkinImport)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "name": "DMHL", "version": "0.1.0"})
	})
	return mux
}

// Start 启动服务
func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.Cfg.ServerPort)
	log.Printf("[DMHL] API 服务启动: http://%s", addr)
	return http.ListenAndServe(addr, s.Routes())
}

func (s *Server) provider() mirror.Provider {
	return mirror.Provider{UseMirror: s.Cfg.UseMirror}
}

// GET /api/versions
func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	manifest, err := mojang.FetchManifest(s.provider())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, manifest)
}

// GET /api/version/detail?url=...
func (s *Server) handleVersionDetail(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get("url")
	if u == "" {
		writeError(w, 400, "缺少 url 参数")
		return
	}
	vd, err := mojang.FetchVersionDetail(u, s.provider())
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, vd)
}

// GET /api/instances
func (s *Server) handleInstances(w http.ResponseWriter, r *http.Request) {
	list, err := s.InstMgr.List()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []*instance.Instance{}
	}
	writeJSON(w, list)
}

// POST /api/instances/create
func (s *Server) handleCreateInstance(w http.ResponseWriter, r *http.Request) {
	var ins instance.Instance
	if err := json.NewDecoder(r.Body).Decode(&ins); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if ins.ID == "" {
		ins.ID = fmt.Sprintf("inst_%d", os.Getpid())
	}
	if err := s.InstMgr.Create(&ins); err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, ins)
}

// POST /api/instances/launch
func (s *Server) handleLaunch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	ins, err := s.InstMgr.Get(req.ID)
	if err != nil {
		writeError(w, 404, "实例不存在")
		return
	}
	// 自动找 Java（缺则下载）
	fj, err := java.FindOrDownload(0)
	if err != nil {
		writeError(w, 500, "Java 准备失败: "+err.Error())
		return
	}
	writeJSON(w, map[string]any{
		"ok":      true,
		"java":    fj.Path,
		"message": fmt.Sprintf("实例 %s 已就绪，Java: %s", ins.Name, fj.Path),
	})
}

// GET/PUT /api/config
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, s.Cfg)
		return
	}
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	s.Cfg = &cfg
	_ = config.Save(configPath(), &cfg)
	writeJSON(w, cfg)
}

// POST /api/auth/offline  { "username": "Steve" }
func (s *Server) handleOfflineLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Username == "" {
		writeError(w, 400, "缺少 username")
		return
	}
	acc := auth.OfflineLogin(req.Username)
	// 追加到账号列表
	list, _ := s.AuthStore.Load()
	list = append(list, acc)
	_ = s.AuthStore.Save(list)
	writeJSON(w, acc)
}

// GET /api/accounts
func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	list, _ := s.AuthStore.Load()
	if list == nil {
		list = []*auth.Account{}
	}
	writeJSON(w, list)
}

// POST /api/skin/import  { "path": "/tmp/x.png", "uuid": "..." }
func (s *Server) handleSkinImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
		UUID string `json:"uuid"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	out, err := s.SkinMgr.ImportPNG(req.Path, req.UUID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": out})
}

// helpers
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	writeJSON(w, map[string]any{"error": msg})
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dmhl", "config.json")
}
