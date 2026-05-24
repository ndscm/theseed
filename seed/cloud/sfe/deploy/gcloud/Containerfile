FROM ghcr.io/ndscm/seed-cloud-sfe-container:latest

CMD [ \
  "--http_port", "8080", \
  "--session_id_cookie_name", "SFESID", \
  "--session_provider", "sql", \
  "--sqlsession_database", "postgres:///sfe_prod?host=/cloudsql/ndscm-prod:us-west1:ndscm-cloudsql-prod&sslmode=disable", \
  "--sqlsession_database_login", "sfe_prod", \
  "--sqlsession_database_secret_file", "/etc/secret/SFE_DATABASE_SECRET/latest", \
  "--golink_openid_discovery", "https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration", \
  "--golink_openid_client_id", "golink-prod", \
  "--golink_openid_client_secret_file", "/etc/secret/GOLINK_OPENID_CLIENT_SECRET/latest", \
  "--golink_service_server", "https://seed-devprod-golink-prod-58770640413.us-west1.run.app", \
  "--stuff_openid_discovery", "https://account.ndscm.com/realms/ndscm/.well-known/openid-configuration", \
  "--stuff_openid_client_id", "stuff-prod", \
  "--stuff_openid_client_secret_file", "/etc/secret/STUFF_OPENID_CLIENT_SECRET/latest", \
  "--stuff_service_server", "https://seed-office-stuff-prod-58770640413.us-west1.run.app", \
  "--verbose" \
  ]
