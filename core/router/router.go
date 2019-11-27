package router

import (
	"github.com/gin-gonic/gin"
	"taotie/core/controllers"
)

type HttpHandle struct {
	Name   string
	Func   gin.HandlerFunc
	Method []string
	Admin  bool
}

var (
	POST = []string{"POST"}
	GET  = []string{"GET"}
	GP   = []string{"POST", "GET"}
)

var (
	HomeRouter = map[string]HttpHandle{
		"/":                     {"Home", controllers.Home, GP, false},
		"/user/token/get":       {"User Token get", controllers.Login, GP, false},
		"/user/token/refresh":   {"User Token refresh", controllers.Refresh, GP, false},
		"/user/token/delete":    {"User Token delete", controllers.Logout, GP, false},
		"/user/register":        {"User Register", controllers.RegisterUser, GP, false},
		"/user/activate":        {"User Verify Email To Activate", controllers.ActivateUser, GP, false},               // 用户自己激活
		"/user/activate/code":   {"User Resend Email Activate Code", controllers.ResendActivateCodeToUser, GP, false}, // 激活码过期重新获取
		"/user/password/forget": {"User Forget Password Gen Code", controllers.ForgetPasswordOfUser, GP, false},       // 忘记密码，验证码发往邮箱
		"/user/password/change": {"User Change Password", controllers.ChangePasswordOfUser, GP, false},                // 根据邮箱验证码修改密码
	}

	// /v1/user/create
	// need login group auth
	V1Router = map[string]HttpHandle{
		// 用户组操作
		"/group/create":        {"Create Group", controllers.CreateGroup, POST, true},
		"/group/update":        {"Update Group", controllers.UpdateGroup, POST, true},
		"/group/delete":        {"Delete Group", controllers.DeleteGroup, POST, true},
		"/group/take":          {"Take Group", controllers.TakeGroup, GP, true},
		"/group/list":          {"List Group", controllers.ListGroup, GP, true},
		"/group/user/list":     {"Group List User", controllers.ListGroupUser, GP, true},         // 超级管理员列出组下的用户
		"/group/resource/list": {"Group List Resource", controllers.ListGroupResource, GP, true}, // 超级管理员列出组下的资源

		// 用户操作
		"/user/list":         {"User List All", controllers.ListUser, GP, true},              // 超级管理员列出用户列表
		"/user/create":       {"User Create", controllers.CreateUser, GP, true},              // 超级管理员创建用户，默认激活
		"/user/assign":       {"User Assign Group", controllers.AssignGroupToUser, GP, true}, // 超级管理员给用户分配用户组
		"/user/update":       {"User Update Self", controllers.UpdateUser, GP, false},        // 更新自己的信息
		"/user/admin/update": {"User Update Admin", controllers.UpdateUserAdmin, GP, true},   // 管理员修改其他用户信息，可以修改用户密码，以及将用户加入黑名单，禁止使用等
		"/user/info":         {"User Info Self", controllers.TakeUser, GP, false},            // 获取自己的信息

		// 资源操作
		"/resource/list":   {"Resource List All", controllers.ListResource, GP, true},              // 列出资源
		"/resource/assign": {"Resource Assign Group", controllers.AssignResourceToGroup, GP, true}, // 资源分配给组

		// 文件操作
		"/file/upload":       {"File Upload", controllers.UploadFile, POST, false},
		"/file/list":         {"File List Self", controllers.ListFile, POST, false},
		"/file/admin/list":   {"File List All", controllers.ListFileAdmin, POST, true}, // 管理员查看所有文件
		"/file/update":       {"File Update Self", controllers.UpdateFile, POST, false},
		"/file/admin/update": {"File Update All", controllers.UpdateFileAdmin, POST, true}, // 管理员修改文件

		"/message/list": {"List Your Message Can Include Private Message", controllers.TakeUser, GP, false},

		"/aws/task/category/add":    {"Add Aws Category Task", controllers.AwsAddCategoryTask, GP, true},
		"/aws/task/category/update": {"Update Aws Category Task, change info or delete", controllers.AwsUpdateCategoryTask, GP, true},
		"/aws/task/category/list":   {"List Aws Category Task", controllers.AwsListCategoryTask, GP, true},
		"/aws/task/asin/add":        {"Add Asin Task", controllers.AwsAddAsinTask, GP, true},
		"/aws/task/asin/update":     {"Update Aws Asin Task, change info or delete", controllers.AwsUpdateAsinTask, GP, true},
		"/aws/task/asin/list":       {"List Aws Asin Task", controllers.AwsListAsinTask, GP, true},
		"/aws/asin/lib/list":        {"List Aws Asin In Lib", controllers.AwsListAsinLib, GP, true},
		"/aws/asin/lib/update":      {"Update Aws Asin In Lib", controllers.AwsUpdateAsinLib, GP, true},
		"/aws/asin/detail/list":     {"List Aws Detail Asin", controllers.AwsListAsinDetail, GP, true},
		"/aws/asin/detail/update":   {"Update Aws Detail Asin", controllers.AwsUpdateAsinDetail, GP, true},
		"/aws/statistics/list":      {"Aws Statistics History list", controllers.AwsListStatistics, GP, true},
		"/aws/task/category/run":    {"Run Aws Category Task Right Now", controllers.AwsRunCategoryTask, GP, true},
		"/aws/task/asin/run":        {"Run Asin Task Right Now", controllers.AwsRunAsinTask, GP, true},
		//"/aws/search/asin":          {"Search Asin", controllers.AwsSearchAsin, GP, true},
	}
)

func SetRouter(router *gin.Engine) {
	for url, app := range HomeRouter {
		for _, method := range app.Method {
			router.Handle(method, url, app.Func)
		}
	}
}

func SetAPIRouter(router *gin.RouterGroup, handles map[string]HttpHandle) {
	for url, app := range handles {
		for _, method := range app.Method {
			router.Handle(method, url, app.Func)
		}
	}
}
