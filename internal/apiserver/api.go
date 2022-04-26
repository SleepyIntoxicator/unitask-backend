package apiserver

import (
	"errors"
	"github.com/gorilla/handlers"
	"net/http"
)

func (s *server) configureUnauthorizedHandlers() {

}

func (s *server) configureRouterAPIv1() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedHeaders([]string{"*"})))

	s.router.HandleFunc("/ping", s.handlePong()).Methods("GET")
	s.router.HandleFunc("/health", s.handleHealth()).Methods("GET")
	s.router.HandleFunc("/api/v1/ping", s.handlePong()).Methods("GET")
	s.router.HandleFunc("/api/v1/logs", s.sendStdoutHandler()).Methods("GET")

	// //=		[__ 	without app authentication		__]
	s.router.Path("/api/v1/auth/app/register").Handler(s.handleAppRegister()).Methods("POST")
	s.router.Path("/api/v1/auth/app/token").Handler(s.handleAppAuthorization()).Methods("GET")

	//<domain>/api
	api := s.router.PathPrefix("/api").Subrouter()
	{

		api.Use(s.authenticateApp)

		s.router.HandleFunc("/api/v1/test/users", s.handleUsers()).Methods("GET")
		s.router.HandleFunc("/api/v1/test/tasks", s.handleTasks()).Methods("GET")

		// //=		[__ 	app authentication required		__]
		auth := api.PathPrefix("/v1/auth").Subrouter()
		{
			auth.HandleFunc("/app/delete", s.handleAppDelete()).Methods("DELETE")
			auth.HandleFunc("/register", s.handleUserRegister()).Methods("POST")
			auth.HandleFunc("/login", s.handleUserSignIn()).Methods("POST")
			auth.HandleFunc("/token", s.handleUserToken()).Methods("POST")
		}

		// //=		[__		user authentication required	__]
		//<domain>/api/v1
		v1 := api.PathPrefix("/v1").Subrouter()
		{
			v1.Use(s.authenticateUser)

			authAuthenticated := v1.PathPrefix("/auth").Subrouter()
			{
				authAuthenticated.HandleFunc("/me", s.handleWhoami()).Methods("GET")
				authAuthenticated.HandleFunc("/logout", s.handleUserLogout()).Methods("GET")
			}

			account := v1.PathPrefix("/account").Subrouter()
			{
				account.HandleFunc("/emailconfimation", s.handleNotImplemented()).Methods("GET")
			}

			////= == == == == == == == == == == == == == == ==//
			//					   ADMIN
			////= == == == == == == == == == == == == == == ==//

			admin := v1.PathPrefix("/admin").Subrouter()
			{
				admin.Use(s.authorizeAdministrator)
				admin.HandleFunc("/users", s.handleUsers()).Methods("GET")
				admin.HandleFunc("/groups", s.handleGroups()).Methods("GET")
				admin.HandleFunc("/tasks", s.handleTasks()).Methods("GET")
				admin.HandleFunc("/tasksV2", s.handleTasksV2()).Methods("GET")
				admin.HandleFunc("/logs", s.sendStdoutHandler()).Methods("GET")
			}

			////= == == == == == == == == == == == == == == ==//
			//					   USERS
			////= == == == == == == == == == == == == == == ==//
			users := v1.PathPrefix("/users").Subrouter()
			{
				users.HandleFunc("/{id:[0-9]+}", s.handleUser()).Methods("GET")
			}

			////= == == == == == == == == == == == == == == ==//
			//					   GROUPS
			////= == == == == == == == == == == == == == == ==//

			groups := v1.PathPrefix("/groups").Subrouter()
			{
				//groups.HandleFunc("", s.handleGroups()).Methods("GET")
				groups.HandleFunc("/{id:[0-9]+}", s.handleGroup()).Methods("GET")
				groups.HandleFunc("/create", s.handleGroupCreate()).Methods("POST")
				groups.HandleFunc("/update", s.handleNotImplemented()).Methods("PUT")
				//	Requires: The user must be a member of the group
				groups.HandleFunc("/{id:[0-9]+}/delete", s.handleGroupDelete()).Methods("DELETE")
				groups.HandleFunc("/{id:[0-9]+}/tasks/create", s.handleCreateGroupTask()).Methods("POST")
				//	Requires: The user must be a member of the group
				groups.HandleFunc("/{id:[0-9]+}/tasks", s.handleGetGroupTasks()).Methods("GET")
				groups.HandleFunc("/{id:[0-9]+}/tasks/{taskId:[0-9]+}", s.handleGetGroupTask()).Methods("GET")
				groups.HandleFunc("/{id:[0-9]+}/members", s.handleGetGroupMembers()).Methods("GET")
				// Require: The user must be a member of the group
				groups.HandleFunc("/{id:[0-9]+}/invite/create", s.handleGroupCreateInvitation()).Methods("GET")
				groups.HandleFunc("/member", s.handleGroupWhereUserIsMember()).Methods("GET")

				v1.HandleFunc("/invite/{hash}", s.handleJoinToGroupWithInvite()).Methods("GET")
			}

			////= == == == == == == == == == == == == == == ==//
			//					   SUBJECTS
			////= == == == == == == == == == == == == == == ==//

			subjects := v1.PathPrefix("/subjects").Subrouter()
			{
				subjects.HandleFunc("", s.handleSubjects()).Methods("GET")
				subjects.HandleFunc("/{id:[0-9]+}", s.handleSubject()).Methods("GET")
				subjects.HandleFunc("/create", s.handleSubjectCreate()).Methods("POST")
				subjects.HandleFunc("/delete/{id:[0-9]+}", s.handleDeleteSubject()).Methods("DELETE")
			}

			////= == == == == == == == == == == == == == == ==//
			//					   TASKS
			////= == == == == == == == == == == == == == == ==//

			tasks := v1.PathPrefix("/tasks").Subrouter()
			{
				tasks.HandleFunc("", s.handleGetAllUserTasks()).Methods("GET")
				tasks.HandleFunc("/{id:[0-9]+}", s.handleGetTask()).Methods("GET")
				tasks.HandleFunc("/personal", s.handleGetUserTasks()).Methods("GET")
				tasks.HandleFunc("/local", s.handleGetUserLocalTasks()).Methods("GET")
				tasks.HandleFunc("/local/create", s.handleCreateUserTask()).Methods("POST")
				tasks.HandleFunc("/create", s.handleCreateGroupTask()).Methods("POST")
				//	?	deprecated
				tasks.HandleFunc("/get/between", s.handleNotImplemented()).Methods()
				//TODO Trello: in_sprint
				tasks.HandleFunc("{id}/assign", s.handleNotImplemented()).Methods("POST")
			}

		}
	}
}

func (s *server) handleNotImplemented() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Warn("handler do nothing")

		s.error(w, r, http.StatusNotImplemented, errors.New("the method has not been implemented yet"))
	}
}

func (s *server) handlePong() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, "pong")
	}
}

func (s *server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, nil)
	}
}



/*

	[__ without app authentication		__]

	/api/v1/auth/app/register
	/api/v1/auth/app/token?app_uuid=$ & app_secret
	/api/v1/auth/app/delete

	[__ app authentication required		__]

	/api/v1/auth/token POST		//if authenticated
	/api/vi/auth/refresh_token POST
	/api/v1/auth/register
	/api/v1/auth/login

	[__	user authentication required	__]

	/api/v1/users

	/api/v1/account
	/api/v1/account/emailconfirmation

	/api/v1/groups
	/api/v1/group
	/api/v1/group/{id}
	/api/v1/group/create
	/api/v1/group/update
	/api/v1/group/delete
	/api/v1/group/tasks
	/api/v1/group/task/{id}

	/api/v1/subjects
	/api/v1/subject/{id}
	/api/v1/subject/create
	/api/v1/subject/delete/{id}

	/api/v1/tasks
	/api/v1/task/{id}
	/api/v1/task/create
	/api/v1/task/update
	/api/v1/task/set_receivers
	/api/v1/task/close
	/api/v1/task/assign_to/ user/{id} group/{id}

*/
