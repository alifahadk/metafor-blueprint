from metafor.dsl.dsl import Server, Work, Source, Program, DependentCall, Constants
from metafor.analysis.visualize import Visualizer

# Define server processing rate
api = {
    "insert": Work(10, [
        DependentCall(
          "server2", "52", "insert", Constants.CLOSED, 10, 3
        )
    ]),
    "get": Work(10, []),
    "put": Work(20, []),
    "list": Work(2, []),
}


# Configure server parameters: queue size, orbit size, threads
server = Server("52", api, qsize=100, orbit_size=20, thread_pool=1)

s2 = Server("server2", api, 20, 5, 1)

# Define client request behavior
src = Source("client", "insert", 9.5, timeout=3, retries=4)

# Build the request-response system
p = Program("Service52")
p.add_server(server)
p.add_source(src)
p.connect("client", "52")