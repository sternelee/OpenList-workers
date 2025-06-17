package routes

import (
	"net/http"

	"github.com/sternelee/OpenList-workers/workers/db"
	"github.com/sternelee/OpenList-workers/workers/drivers"
	"github.com/sternelee/OpenList-workers/workers/handlers"
)

// SetupRoutes 设置所有路由
func SetupRoutes(repos *db.Repositories, driverService *drivers.DriverService) *http.ServeMux {
	mux := http.NewServeMux()

	// 创建用户驱动服务
	userDriverService := drivers.NewUserDriverService(repos)

	// 创建处理器
	authHandler := handlers.NewAuthHandler(repos)
	driverHandler := handlers.NewDriverHandler(driverService)
	storageHandler := handlers.NewStorageHandler(repos, driverService)
	userStorageHandler := handlers.NewUserStorageHandler(repos, userDriverService)
	userFileHandler := handlers.NewUserFileHandler(userDriverService)

	// 认证相关路由
	mux.HandleFunc("/api/auth/login", corsMiddleware(authHandler.Login))
	mux.HandleFunc("/api/auth/register", corsMiddleware(authHandler.Register))
	mux.HandleFunc("/api/auth/logout", corsMiddleware(authHandler.Logout))
	mux.HandleFunc("/api/auth/me", corsMiddleware(requireAuth(authHandler.GetCurrentUser, repos)))

	// 驱动相关路由
	mux.HandleFunc("/api/drivers", corsMiddleware(driverHandler.ListDriverNames))
	mux.HandleFunc("/api/drivers/info", corsMiddleware(driverHandler.ListDriverInfo))
	mux.HandleFunc("/api/drivers/info/", corsMiddleware(driverHandler.GetDriverInfo))

	// 用户存储管理路由 (需要用户认证)
	mux.HandleFunc("/api/user/storages", corsMiddleware(requireAuth(userStorageHandler.ListUserStorages, repos)))
	mux.HandleFunc("/api/user/storages/create", corsMiddleware(requireAuth(userStorageHandler.CreateUserStorage, repos)))
	mux.HandleFunc("/api/user/storages/update", corsMiddleware(requireAuth(userStorageHandler.UpdateUserStorage, repos)))
	mux.HandleFunc("/api/user/storages/delete", corsMiddleware(requireAuth(userStorageHandler.DeleteUserStorage, repos)))
	mux.HandleFunc("/api/user/storages/test", corsMiddleware(requireAuth(userStorageHandler.TestUserStorage, repos)))

	// 用户文件操作路由 (需要用户认证)
	mux.HandleFunc("/api/user/fs/list", corsMiddleware(requireAuth(userFileHandler.ListUserFiles, repos)))

	// 管理员存储管理路由 (需要管理员权限)
	mux.HandleFunc("/api/admin/storages", corsMiddleware(requireAuth(requireAdmin(storageHandler.ListStorages, repos), repos)))
	mux.HandleFunc("/api/admin/storages/create", corsMiddleware(requireAuth(requireAdmin(storageHandler.CreateStorage, repos), repos)))
	mux.HandleFunc("/api/admin/storages/update", corsMiddleware(requireAuth(requireAdmin(storageHandler.UpdateStorage, repos), repos)))
	mux.HandleFunc("/api/admin/storages/delete", corsMiddleware(requireAuth(requireAdmin(storageHandler.DeleteStorage, repos), repos)))
	mux.HandleFunc("/api/admin/storages/enable", corsMiddleware(requireAuth(requireAdmin(storageHandler.EnableStorage, repos), repos)))
	mux.HandleFunc("/api/admin/storages/disable", corsMiddleware(requireAuth(requireAdmin(storageHandler.DisableStorage, repos), repos)))
	mux.HandleFunc("/api/admin/storages/test", corsMiddleware(requireAuth(requireAdmin(storageHandler.TestStorage, repos), repos)))

	// 公开文件访问路由 (支持匿名和用户访问)
	mux.HandleFunc("/d/", corsMiddleware(userFileHandler.GetUserFileLink))
	mux.HandleFunc("/download/", corsMiddleware(userFileHandler.DownloadUserFile))

	// 兼容性文件操作路由 (管理员使用)
	mux.HandleFunc("/api/fs/list", corsMiddleware(requireAuth(requireAdmin(driverHandler.ListFiles, repos), repos)))
	mux.HandleFunc("/api/fs/get", corsMiddleware(requireAuth(requireAdmin(driverHandler.GetFile, repos), repos)))

	// 健康检查
	mux.HandleFunc("/health", corsMiddleware(healthCheck))

	return mux
}

// healthCheck 健康检查处理器
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"openlist-workers"}`))
}

// corsMiddleware CORS中间件
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// requireAuth 需要认证的中间件
func requireAuth(next http.HandlerFunc, repos *db.Repositories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取Authorization头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrorResponse(w, 401, "authorization header required")
			return
		}

		// 检查Bearer token格式
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			writeErrorResponse(w, 401, "invalid authorization header format")
			return
		}

		token := authHeader[7:]

		// 验证token
		claims, err := handlers.VerifyJWT(token)
		if err != nil {
			writeErrorResponse(w, 401, "invalid token")
			return
		}

		// 获取用户信息
		user, err := repos.User.GetByID(r.Context(), claims.UserID)
		if err != nil {
			writeErrorResponse(w, 401, "user not found")
			return
		}

		if user.Disabled {
			writeErrorResponse(w, 401, "user disabled")
			return
		}

		// 将用户信息添加到context
		ctx := handlers.SetUserInContext(r.Context(), user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

// requireAdmin 需要管理员权限的中间件
func requireAdmin(next http.HandlerFunc, repos *db.Repositories) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := handlers.GetUserFromContext(r.Context())
		if user == nil {
			writeErrorResponse(w, 401, "user not found in context")
			return
		}

		if !user.IsAdmin() {
			writeErrorResponse(w, 403, "admin permission required")
			return
		}

		next(w, r)
	}
}

// writeErrorResponse 写错误响应
func writeErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	// 简单的JSON编码
	if code == 401 {
		w.Write([]byte(`{"code":401,"message":"` + message + `"}`))
	} else if code == 403 {
		w.Write([]byte(`{"code":403,"message":"` + message + `"}`))
	} else {
		w.Write([]byte(`{"code":` + string(rune(code)) + `,"message":"` + message + `"}`))
	}
}

