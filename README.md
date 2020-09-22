# choo-pos-backend

## Note
- Repo: https://github.com/choobot/choo-pos-backend
- API Doc: https://choo-pos-backend.herokuapp.com

## Live Demo
- https://choo-pos-backend.herokuapp.com

## Prerequisites for Development
- Mac or Linux which can run shell script
- Docker
- ngrok CLI
- Heroku CLI (for Production Deployment only)

## Local Running and Expose to the internet
- Create LINE Login channel in LINE Developers Console
- $ ./local-dev-server.sh
- Config environment variables in .env (see example in .env.example)
- The Callback URL for LINE Login API will be something like https://cebccce9ede4.ngrok.io/auth config it in LINE Developers Console
- $ ./dev.sh

## Unit Testing
- Config environment variables in .env (see example in .env.example)
- $ ./test.sh

## Production Deployment
- Create LINE Login channel in LINE Developers Console
- Create Heroku App with ClearDB MySQL add-on in Heroku Dashboard
- Config environment variables in .env (see example in .env.example)
- The Callback URL for LINE Login API will be something like https://choo-pos-backend.herokuapp.com/auth config it in LINE Developers Console
- $ ./prod.sh

## Postman Testing

## Tech Stack
- Go
- Echo
- GoMock
- OAuth 2.0
- LINE Login API
- MySQL
- Docker
- Heroku
- Swagger
- Postman