Ok so we plan on running multiple services :
-frontend(whatever framework we will use)
-Backend API(in go lang)
-Websocket service(go lang)
-Database(SqLite)

This helps in service isolation (each one runs in its own container),we have the same environment everywhere therefore -> we package the project once and it runs everywhere,we all got the same setup.

E.g. now manos has a port listening on server start at 8081.Up till now,every exercise we did,it was in this same manner :

func main() {
    db, err := OpenDB("social_network.db")
    // Single server on port 8081
    http.ListenAndServe(":8081", nil)
    blablabla
}

Instead of our old monolithic big ass mofo approach,we will now have different ports for each service :
-Auth Service (Port 8081)
-User Service (Port 8082)
-Post Service (Port 8083)
-Group Service(Port 8084)
-Chat/Websocket Service (Port 8085)
-DataBase Service (our sqlite with migrations)

so the project architecture will look something like this  : 

social-network/
├── docker-compose.yml
├── 
├── go.mod                  # Your existing file
├── services/
│   ├── auth/
│   │   ├── Dockerfile
│   │   └── main.go
│   ├── users/
│   │   ├── Dockerfile
│   │   └── main.go
│   ├── posts/
│   │   ├── Dockerfile
│   │   └── main.go
│   ├── groups/
│   │   ├── Dockerfile
│   │   └── main.go
│   └── chat/
│       ├── Dockerfile
│       └── main.go
├── frontend/
│   ├── Dockerfile
│   └── [Next.js files]
└── db/
    └── migrations/         # Your existing migrations

For any GO microservice,the dockerfile will probably look the same(define the workdir,copy the go.mod,go.sum,run go mod download, copy . .  and so on,you know the drill.)

Now the juice : 
Containers cummunicate through Docker Networks :
Example what the docker-compose.yml will contain :

# docker-compose.yml
version: '3.8'

networks:
  social-network:
    driver: bridge

services:
  database:
    image: alpine:latest
    volumes:
      - ./social_network.db:/app/social_network.db
      - ./db/migrations:/app/migrations
    networks:
      - social-network

  auth-service:
    build: ./services/auth
    ports:
      - "8081:8081"
    depends_on:
      - database
    networks:
      - social-network
    environment:
      - DB_PATH=/app/social_network.db

  user-service:
    build: ./services/users
    ports:
      - "8082:8082"
    depends_on:
      - database
      - auth-service
    networks:
      - social-network

  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    depends_on:
      - auth-service
      - user-service
    networks:
      - social-network

The communication pattern goes as such : 
      HTTP API Calls :
      Frontend is in constant communication with the Backend Services,
      each service communicates with other services

      Database Sharing:
      ALL services connect to the same SQLite databse,using the OpenDB function we got

      WebSocket connections:
      These are real time features and are persistent connections
      

For example,we now have 2 seperate services,the auth-service and the user-service as well ass a dockerfile for each one and an extended docker-compose.yml file.Each dockerfile is the blueprint,the recipe on how to build the service,with what dependencies and so on.We essentially lock our version so its usable everywhere:
1. takes auth service
2. compiles it(into binary)
3. creates a container image
4. when it runs,it runs the auth service binary on the specified port

Same thing happens for the user-service.

Now the yml orchastrates all the containers :
1. Creates a network so the containers can communicate(see the file it says networks : social-network(project) driver : bridge->it bridges them (WOW!!!!!!!))

2. Defines each service and instructs the dockerfiles to build each service along with each port(maps the host port to the container port).Important.The host port is the port that will run on our computer.Dont confuse it with the port that the container will run on.The host port is actually the one that allows the containers to communicate outside of their scope.It acts as the gateway.Containers talk to each other using the containers names + the container ports.(In code)

3. Defines dependencies,because we must tell it for each service : huh...what NEEDS to run before this so it can actually work???For example the user service needs both the database AND the auth-service to run.

4. Adds the container on the social-network(defined before) network.

5. then we define the env variables,the configuration settings we pass to the containers.Just like we had a config file in graphQL project,i think they help both abstract and have the same name for different environments.

6. And then we got the volumes(Shared folders between our computer and the containers)
- Example : 
        volumes:
        - "./social_network.db:/app/social_network.db"
        #      ↑                      ↑
        #   HOST PATH            CONTAINER PATH

 This helps to not lose data from the database when we run it inside of containers : 

 # Without Volumes
 ┌────────────────────────────────────────────────────────────┐
│  SCENARIO: Working without volumes                          │
│                                                             │
│  1. docker-compose up                                       │
│  2. 👤 Create 100 users, 500 posts, 50 groups               │
│  3. 💥 Container crashes / docker-compose down              │
│  4. docker-compose up                                       │
│  5. 😱 ALL YOUR WORK IS GONE!                               │
│                                                             │
│  Database lived INSIDE the container                        │
│  Container dies = Database dies                             │
└─────────────────────────────────────────────────────────────┘    

# With Volumes
┌─────────────────────────────────────────────────────────────┐
│  SCENARIO: Working with volumes (Your setup)                │
│                                                             │
│  1. docker-compose up                                       │
│  2. 👤 Create 100 users, 500 posts, 50 groups               │
│  3. 💥 Container crashes / docker-compose down              │
│  4. docker-compose up                                       │
│  5. 😊 ALL YOUR DATA IS STILL THERE!                        │
│                                                             │
│  Database lives on YOUR COMPUTER                            │
│  Container dies = Database survives                         │
└─────────────────────────────────────────────────────────────┘
Our database file social_network.db lives on our computer,all containers just "borrow" access to it
If containers crash, your database survives on your computer