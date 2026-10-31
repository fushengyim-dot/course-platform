from pathlib import Path
import json, subprocess, time, urllib.request, urllib.error
import argparse
parser=argparse.ArgumentParser(description='Verify a dedicated, disposable local Compose stack.')
parser.add_argument('--env-file',required=True)
parser.add_argument('--port',type=int,default=19871)
args=parser.parse_args()
base='http://127.0.0.1:'+str(args.port)
# Fixed project identity prevents accidental operation of a default project.
cli=['docker','compose','-p','foundation-independent-runtime','--env-file',args.env_file]
op=urllib.request.build_opener(urllib.request.ProxyHandler({}))
results=[]
def compose(*args,ok=True):
 r=subprocess.run(cli+list(args),capture_output=True,text=True,encoding='utf-8');
 if ok and r.returncode:raise RuntimeError('compose '+args[0]+' failed')
 return r
def request(path,status,code=None):
 start=time.monotonic()
 try:r=op.open(base+path,timeout=4)
 except urllib.error.HTTPError as e:r=e
 with r: body=r.read().decode();actual=r.status
 assert actual==status,(path,actual,status)
 elapsed=time.monotonic()-start
 obj=json.loads(body) if code else None
 if code:assert obj['code']==code and obj['request_id']
 results.append({'path':path,'status':status,'code':code,'seconds':round(elapsed,3)})
 return obj,elapsed
def recover():
 for _ in range(40):
  try:
   h,_=request('/health',200,'OK');assert all(x=='up' for x in h['data']['dependencies'].values());return
  except (AssertionError,urllib.error.URLError):time.sleep(1)
 raise RuntimeError('recovery failed')
request('/',200);request('/status/details',200);request('/unknown',404,'NOT_FOUND');request('/api/v1/records',401,'UNAUTHORIZED');recover()
for dep in ['mysql','qdrant']:
 try:
  compose('stop',dep)
  h,elapsed=request('/health',503,'DEPENDENCY_UNAVAILABLE');assert h['data']['dependencies'][dep]=='down' and elapsed<3
  request('/live',200,'OK')
 finally:compose('start',dep)
 recover()
compose('up','-d','--no-deps','--force-recreate','api');recover()
# A running API with an intentionally wrong password must answer a real probe.
# One-off container shares only this independent network and has no host ports.
name='foundation-independent-wrong-password'
try:
 compose('run','-d','--no-deps','--name',name,'-e','MYSQL_PASSWORD=deliberately-invalid','api')
 command=['docker','run','--rm','--network','foundation-independent-runtime_default','foundation-independent-check-checks','node','-e',"fetch('http://foundation-independent-wrong-password:8080/health').then(async r=>{const b=await r.json();if(r.status!==503||b.code!=='DEPENDENCY_UNAVAILABLE'||b.data.dependencies.mysql!=='down'||b.data.dependencies.qdrant!=='up')process.exit(1);console.log('wrong password: actual HTTP 503, mysql down, qdrant up')}).catch(()=>process.exit(2))"]
 r=subprocess.run(command,capture_output=True,text=True);assert r.returncode==0,r.stderr
 results.append({'check':'wrong password on running API','result':'PASS'})
finally:subprocess.run(['docker','rm','-f',name],capture_output=True)
r=compose('run','--rm','--no-deps','-e','APP_VERSION=','api',ok=False);assert r.returncode!=0 and 'APP_VERSION' in r.stderr+r.stdout
results.append({'check':'missing configuration exits nonzero','result':'PASS'})
r=compose('logs','--no-color','--no-log-prefix','api');rows=[json.loads(x) for x in r.stdout.splitlines() if x.startswith('{')];assert rows
assert all(all(k in x for k in ['time','level','module','request_id','method','route','status','duration_ms']) for x in rows)
for line in Path(args.env_file).read_text().splitlines():
 if line.startswith(('MYSQL_PASSWORD=','MYSQL_ROOT_PASSWORD=')):assert line.split('=',1)[1] not in r.stdout
results.append({'check':'request log schema and generated secret absence','result':'PASS'})
print(json.dumps(results,ensure_ascii=False))
