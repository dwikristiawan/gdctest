package modules

import (
	"gdctest/internal/domain/task/controller"
	"gdctest/internal/domain/task/repository"
	"gdctest/internal/domain/task/service"
	"gdctest/pkg/infra"
)

func init() {
	Register(func(ctx *Context) error {
		db, _ := infra.GetDB(ctx.Infra)
		rdb, _ := infra.GetRedis(ctx.Infra)
		repo := repository.NewPostgresTaskRepository(db)
		svc := service.NewTaskService(repo, rdb)
		ctrl := controller.NewTaskController(svc)

		// Teams endpoints (/api/v1.0/teams)
		teams := ctx.Echo.Group("/v1.0/teams")
		if ctx.AuthMiddleware != nil {
			teams.Use(ctx.AuthMiddleware)
		}
		teams.POST("", ctrl.CreateTeam)
		teams.GET("", ctrl.GetTeams)
		teams.DELETE("/:id", ctrl.DeleteTeam)
		teams.POST("/:id/members", ctrl.AddMember)
		teams.GET("/:id/members", ctrl.GetTeamMembers)
		teams.DELETE("/:id/members/:user_id", ctrl.RemoveMember)

		// Tasks endpoints (/api/v1.0/tasks)
		tasks := ctx.Echo.Group("/v1.0/tasks")
		if ctx.AuthMiddleware != nil {
			tasks.Use(ctx.AuthMiddleware)
		}
		tasks.POST("", ctrl.CreateTask)
		tasks.GET("", ctrl.GetTasks)
		tasks.GET("/:id", ctrl.GetTaskDetail)
		tasks.PUT("/:id", ctrl.UpdateTask)
		tasks.DELETE("/:id", ctrl.DeleteTask)
		tasks.POST("/:id/assign", ctrl.AssignTask)

		ctx.RegisterService("task_service", svc)
		return nil
	})
}
