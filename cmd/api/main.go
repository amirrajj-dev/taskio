package main

import "github.com/amirrajj-dev/taskio/internal/app"



// @title           Taskio API
// @version         1.0
// @description     Task management system API with real-time features
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@taskio.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:3000
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in                         cookie
// @name                       taskio-cookie
// @description                JWT token stored in cookie
func main() {
	app.Bootstrap()
}
