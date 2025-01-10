DESCRIPTION:
A simple tcp based message feed written in go , that allows clients to publish messages to a message feed in near real time.

HOW TO RUN
- make app

```
% make app
Running setup on macOS. Checking for tmux...
Setup complete. You can now run 'make start-db' or 'make serve' etc.
Recreating database container from scratch...
docker stop go_chat_db || true
Error response from daemon: No such container: go_chat_db
docker rm go_chat_db || true
Error response from daemon: No such container: go_chat_db
docker volume rm go_chat_volume || true
Error response from daemon: get go_chat_volume: no such volume
docker run -d --name go_chat_db \
                -e POSTGRES_PASSWORD=postgres \
                -e POSTGRES_USER=postgres \
                -e POSTGRES_DB=go-chat \
                -v go_chat_volume:/var/lib/postgresql/data \
                -p 5432:5432 \
                postgres:15
0b1f5759802bed5dac8b10c3376f4306744db93562d56e548859e86e9e5f8f6f
Database container (re)created successfully.
App successfully started with all components running.
% 
```
```
% tmux ls
feed: 1 windows (created Fri Jan 10 21:03:45 2025)
publisher: 1 windows (created Fri Jan 10 21:03:45 2025)
server: 1 windows (created Fri Jan 10 21:03:43 2025)
% 

```

**publisher**

```
2025/01/10 21:43:11 PostgreSQL connected , tables created
name (atleast 4 characters): << john
2025/01/10 21:44:37 userID-john
<< hello world my name is john  😄      
<< happy to be here 💯 🔥😎
<<                                                                         
```
**feed**

```
2025/01/10 21:43:11 PostgreSQL connected , tables created
john >> hello world my name is jon 😄 [1/10/2025 18:45:01]
john >> happy to be here 💯🔥😎 [1/10/2025 18:45:26]
```

**server**

```
2025/01/10 21:43:09 PostgreSQL connected , tables created
[readMessages] 2025/01/10 21:43:09 server is accepting connections on 3000
[writeMessages]  2025/01/10 21:43:09 server is accepting connections on 2000
[readMessages] 2025/01/10 21:43:11 new connection received
[writeMessages]  2025/01/10 21:44:37 user successfully created: &{1 john}
```



**TERMINATE**

- make stop


**INSPIRED BY**:

https://codingchallenges.fyi/challenges/challenge-realtime-chat