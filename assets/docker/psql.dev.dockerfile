FROM postgres:12

ADD ./assets/sql/postgresql/bootstrap/* /docker-entrypoint-initdb.d/
