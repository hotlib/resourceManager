#! /bin/sh

set +x

go build -o ./entCmd ./graph/cmd/entscript/
go build -o ./resourceManager ./graph/cmd/graph/

docker run --name mysql -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 -d mysql:latest

echo "Waiting for mysql to initialize..."
sleep 20

cat <<EOF | docker exec -i mysql bash
echo 'CREATE DATABASE IF NOT EXISTS \`tenant_FRINX\` DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_bin;' > init
mysql -u root -h localhost -P 3306 -proot < init
EOF

./entCmd --tenant FRINX --db-dsn "root:root@tcp(localhost:3306)/?charset=utf8&parseTime=true&interpolateParams=true" --user devel
./resourceManager --mysql.dsn="root:root@tcp(localhost:3306)/?charset=utf8&parseTime=true&interpolateParams=true"  --tenancy.db_max_conn=10 --web.listen-address=0.0.0.0:8884  --grpc.listen-address=0.0.0.0:8885 
