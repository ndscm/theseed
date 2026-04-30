FROM ghcr.io/ndscm/seed-devprod-golink-container:latest

CMD [ \
  "--golink_database", "postgres:///golink_prod?host=/cloudsql/ndscm-prod:us-west1:ndscm-cloudsql-prod&sslmode=disable", \
  "--golink_database_login", "golink_prod", \
  "--golink_database_secret_file", "/etc/secret/GOLINK_DATABASE_SECRET", \
  "--verbose" \
  ]
