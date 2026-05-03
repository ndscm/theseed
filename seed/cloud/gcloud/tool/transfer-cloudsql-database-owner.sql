\set target_database service_prod
\set target_user service_prod

GRANT :"target_user" TO postgres;
ALTER DATABASE :"target_database" OWNER TO :"target_user";
REVOKE :"target_user" FROM postgres;
