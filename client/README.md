An example oidc client implementation.
Provide required Envs for client execution:

* AUTH_URL
* TOKEN_URL
* REDIRECT_URL

like this
AUTH_URL=https://localhost:4444/oauth2/auth TOKEN_URL=https://localhost:4444/oauth2/token REDIRECT_URL=http://localhost:3000/callback go run client/main.go