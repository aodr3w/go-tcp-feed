add db agonostic database layer
add authentication to access messages
add k8s support
add logging support
add size limit for messages


OPTIMIZATIONS
- look into increase buffer size on *net.Conn to handle faster send speed instead of using time.sleep()
