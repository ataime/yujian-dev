docker 部署

```
docker-compose up --build

docker-compose down
```

裸机部署

```
修改数据库连接：
	dataSourceName := dbUser + ":" + dbPassword + "@/" + dbName
	// dataSourceName := fmt.Sprintf("%s:%s@tcp(db:3306)/%s", dbUser, dbPassword, dbName)

运行:
    go run .
```

功能页面

http://X.X.X.X/admin

http://X.X.X.X/user/2
