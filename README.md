# RUN query.sql
# 1 - Make sure your Postgresql is running 
- Install postgis package - `sudo apt install postgis`
- Create extension inside your psql shell - `CREATE EXTENSION IF NOT EXISTS postgis;`

# 2 Run the command
- Run:  `psql -U your_username -d your_database_name -f query.sql`


# Run endpoint.go
- sudo apt install jq
- Make sure you modify the following variables in the `endpoint.go` file according to your database settings:
    const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "dbname"
)
    
- run `go run endpoint.go`
- in another terminal run the test requests
    - curl -X GET 'http://localhost:8000/spots?latitude=40.7128&longitude=-74.0060&radius=1000&type=circle' | jq
    - curl -X GET 'http://localhost:8000/spots?latitude=40.7128&longitude=-74.0060&radius=500&type=circle' | jq
	- curl -X GET 'http://localhost:8000/spots?latitude=40.7128&longitude=-74.0060&radius=5000&type=square' | jq
