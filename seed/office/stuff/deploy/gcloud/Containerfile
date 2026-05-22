FROM ghcr.io/ndscm/seed-office-stuff-container:latest

COPY ./openid_discovery.json /opt/stuff/openid_discovery.json

CMD [ \
  "--port", "8080", \
  "--openid_discovery", "/opt/stuff/openid_discovery.json", \
  "--stuff_database", "postgres:///stuff_prod?host=/cloudsql/ndscm-prod:us-west1:ndscm-cloudsql-prod&sslmode=disable", \
  "--stuff_database_login", "stuff_prod", \
  "--stuff_database_secret_file", "/etc/secret/STUFF_DATABASE_SECRET/latest", \
  "--verbose" \
  ]
