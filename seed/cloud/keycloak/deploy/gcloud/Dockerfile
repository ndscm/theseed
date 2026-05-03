FROM ghcr.io/ndscm/seed-cloud-keycloak-container:latest

ENV KC_DB_URL="jdbc:postgresql:///keycloak_prod?socketFactory=org.newsclub.net.unix.AFUNIXSocketFactory\$FactoryArg&socketFactoryArg=/cloudsql/ndscm-prod:us-west1:ndscm-cloudsql-prod/.s.PGSQL.5432"
ENV KC_DB_USERNAME="keycloak_prod"
# Set the database password with secret manager
# KC_DB_PASSWORD=

# Set the bootstrap admin username and password for the first run.
# KC_BOOTSTRAP_ADMIN_USERNAME=unsecure
# KC_BOOTSTRAP_ADMIN_PASSWORD=unsecure
