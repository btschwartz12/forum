A simle forum with authentication and registration, which can be run through a Cloudflare tunnel.

You can find a live example of this at [forum.btschwartz.com](https://forum.btschwartz.com).

## Usage

First, add some environment variables:
```bash
$ cat .env
FORUM_AUTH_TOKEN=your_auth_token
FORUM_SESSION_KEY=your_session_key
FORUM_SLACK_WEBHOOK_URL=<slack-webhook-url> # optional
```

Now, you can start the server:
```bash
# note that you either need to disable or install the 'swag' and 'sqlc' commands
$ CGO_ENABLED=0 go build -o forum main.go
$ ./forum --port 8080 --public --var-dir var
```

Now, you can access the site at `http://localhost:8080` and the Swagger UI at `http://localhost:8080/api`.

### Auth Setup

First, set yourself up as an admin user. This can be done through the Swagger UI, or by something like this:

```bash
curl -X 'POST' \
  'http://localhost:8000/api/users' \
  -H 'accept: application/json' \
  -H 'Authorization: your_auth_token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'username=user&password=pass&is_admin=true'
```

Now, you will be able to edit/delete posts from any users, as well as add/edit/delete users (through the Swagger UI).

### Cloudflare Tunnel Setup

I'm going to assume that you have already linked your domain to Cloudflare. If not, you can do so [here](https://dash.cloudflare.com/).

First, create a tunnel following [this guide](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/get-started/create-remote-tunnel/). When you are setting up the public hostname, make sure to set the service to be `Type=HTTP` and `URL=forum:8080`.

Now, get the token value and add it to your `.env` file:
```bash
$ cat .env
...
TUNNEL_TOKEN=<your_tunnel_token>
```

Now, you can run the forum through the tunnel:
```bash
$ docker compose up
```

Hopefully, it all works...

