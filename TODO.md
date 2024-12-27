add db agonostic database layer
add authentication to access messages
add k8s support
add logging support


OPTIMIZATIONS
- look into increase buffer size on *net.Conn to handle faster send speed instead of using time.sleep()

BUGS:
- old messages are reloaded when i send new messages to the chart