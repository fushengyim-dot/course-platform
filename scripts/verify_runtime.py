"""Run only against this project's local Compose stack."""
import json,subprocess,time,urllib.request,urllib.error,sys
from pathlib import Path
op=urllib.request.build_opener(urllib.request.ProxyHandler({}))
base=sys.argv[1] if len(sys.argv)>1 else 'http://127.0.0.1:18770'
results=[]
def request(path,status,method='GET'):
 start=time.monotonic()
 try: r=op.open(urllib.request.Request(base+path,method=method),timeout=4)
 except urllib.error.HTTPError as e:r=e
 with r: body=r.read().decode();actual=r.code
 assert actual==status,(path,actual,status)
 elapsed=round(time.monotonic()-start,3)
 obj=json.loads(body)
 assert obj['request_id']
 results.append({'path':path.split('?')[0],'method':method,'status':actual,'code':obj['code'],'seconds':elapsed})
 return obj

def compose(*args,ok=True):
 r=subprocess.run(['docker','compose',*args],capture_output=True,text=True,encoding='utf-8')
 if ok and r.returncode:raise RuntimeError('Compose command failed: '+args[0])
 return r

h=request('/health',200);assert h['data']['dependencies']=={'mysql':'up','qdrant':'up'}
assert request('/api/v1/resources',401)['code']=='UNAUTHORIZED'
assert request('/unknown',404)['code']=='NOT_FOUND'
assert request('/health',405,'POST')['code']=='METHOD_NOT_ALLOWED'
assert request('/health?secret=runtime-private-marker',400)['code']=='INVALID_ARGUMENT'
for dep in ['qdrant','mysql']:
 try:
  compose('stop',dep)
  h=request('/health',503);assert h['data']['dependencies'][dep]=='down'
  assert results[-1]['seconds']<3
  assert request('/live',200)['data']['status']=='alive'
 finally:compose('start',dep)
 for attempt in range(30):
  try:
   h=request('/health',200)
   if all(v=='up' for v in h['data']['dependencies'].values()):break
  except (AssertionError,urllib.error.URLError):time.sleep(1)
 else:raise RuntimeError('dependency did not recover')
results.append({'check':'dependency stop and recovery','result':'PASS'})
compose('up','-d','--force-recreate','--no-deps','api')
for attempt in range(15):
 try:
  request('/health',200);break
 except (AssertionError,urllib.error.URLError):time.sleep(1)
else:raise RuntimeError('gateway did not recover backend replacement')
results.append({'check':'gateway tracks replaced backend','result':'PASS'})
r=compose('run','--rm','--no-deps','-e','APP_VERSION=','api',ok=False)
assert r.returncode!=0 and 'APP_VERSION' in r.stderr+r.stdout
results.append({'check':'missing configuration process exit','result':'PASS'})
r=compose('logs','--no-color','--no-log-prefix','api')
lines=[json.loads(line) for line in r.stdout.splitlines() if line.startswith('{')]
assert lines and all(all(k in row for k in ['time','level','module','request_id','status','duration_ms']) for row in lines)
assert 'runtime-private-marker' not in r.stdout and 'deliberately-invalid' not in r.stdout
# Compare known local generated credentials without printing them.
for line in Path('.env').read_text().splitlines():
 if line.startswith(('MYSQL_PASSWORD=','MYSQL_ROOT_PASSWORD=')):
  secret=line.split('=',1)[1];assert secret not in r.stdout
results.append({'check':'JSON log schema and secret absence','result':'PASS','lines':len(lines)})
print(json.dumps(results,ensure_ascii=False,indent=2))
