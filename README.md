# choo-pos-backend

## Note
- Repo: https://github.com/choobot/choo-pos-backend

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
- The Callback URL for LINE Login API will be something like https://cebccce9ede4.ngrok.io/auth
- Config environment variables in .env (see example in .env.example)
- Config Callback URL in LINE Developers Console
- $ ./dev.sh

## Production Deployment
- Create LINE Login channel in LINE Developers Console
- Create Heroku App in Heroku Dashboard
- The Callback URL for LINE Login API will be something like https://choo-pos-backend.herokuapp.com/auth
- Config environment variables in .env (see example in .env.example)
- Config Callback URL in LINE Developers Console
- $ ./prod.sh