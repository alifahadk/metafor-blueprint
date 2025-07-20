from metafor.dsl.dsl import Server, Work, Source, Program, DependentCall, Constants
from metafor.analysis.visualize import Visualizer

api = { 'insert ': Work (10 , [] ,) }

server = Server ('simple', api, qsize=35, thread_pool=1)
src = Source ('client', 'insert', rate=5, timeout=5, retries=5)


p = Program('SimpleService')
p.add_server(server).add_source(src).connect ('client', 'simple')
