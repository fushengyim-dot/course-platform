"""Local staged/history boundary check; never prints matching content."""
from pathlib import Path
import re,subprocess,sys

def git(*args):return subprocess.check_output(['git',*args])
secrets=[]
if Path('.env').exists():
 for line in Path('.env').read_text().splitlines():
  k,sep,v=line.partition('=')
  if sep and any(x in k for x in ['PASSWORD','SECRET','TOKEN','KEY']) and v:secrets.append(v.encode())
paths=git('diff','--cached','--name-only','--diff-filter=ACMR').decode().splitlines()
errors=[]
def inspect(label,data):
 if any(value in data for value in secrets):errors.append(label+': configured secret found')
 if re.search(rb'-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----',data):errors.append(label+': private key found')
for path in paths:
 if path=='.env' or path.startswith(('internal/','tmp/','data/','uploads/')) or ('/.env' in path and not path.endswith('.env.example')):errors.append(path+': excluded material staged')
 inspect(path,git('show',':'+path))
if '--history' in sys.argv:
 for commit in git('rev-list','--all').decode().splitlines():
  inspect('commit '+commit,git('show','--format=','--binary',commit))
for error in errors:print(error)
print('secret boundary check: '+('FAIL' if errors else 'PASS'))
sys.exit(bool(errors))
