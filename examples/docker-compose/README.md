# Docker Compose

This is an example of running the exporter and [PMS](https://www.plex.tv/media-server-downloads/) within [Docker Compose](https://docs.docker.com/compose/). You can tweak the `docker-compose.yaml` file to suit your needs, or borrow the exporter and/or init container to run on your server.

You'll need two pieces of information - the URL to the server you want to monitor, and a Plex token belonging to the server owner. The URL needs to be relative to where the exporter is running. In this example both run within Docker Compose, so `http://plex:32400` works.

Rather than finding an existing token, you can get a claim token at [plex.tv/claim](https://plex.tv/claim).

With a claim token in hand, you can run the following from this directory:

```sh
PLEX_CLAIM=YOURCLAIMTOKEN docker compose up
```

This will start the server, claim it with your Plex account, and privately share the token with the Prometheus exporter.

One running, you are able to point Prometheus or Grafana Agent at `localhost:9000` to start scraping metrics.