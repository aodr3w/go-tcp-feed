add db agonostic database layer
add authentication to access messages
add k8s support
add logging support


OPTIMIZATIONS
- display messages on client side with time stamp e.g msg text [timestamp] / >> msg text [author, timestamp]
- look into increase buffer size on *net.Conn to handle faster send speed instead of using time.sleep()