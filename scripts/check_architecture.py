from pathlib import Path
import re
import sys
errors=[]
for p in Path('backend').rglob('*.go'):
 if p.name.endswith('_test.go'): continue
 layer=p.parent.name
 text=p.read_text()
 forbidden={'service':['net/http','database/sql','/adapter','/handler'], 'repository':['/service','/handler','/adapter'], 'handler':['database/sql','/adapter','go-sql-driver'], 'adapter':['/service','/handler']}.get(layer,[])
 for value in forbidden:
  if value in text: errors.append(f'{p}: forbidden dependency {value}')
 if layer=='service' and re.search(r'(?i)\b(select|insert|update|delete)\s+.*\b(from|into|set)\b',text): errors.append(f'{p}: SQL in service')
 if 'internal' in p.parts and '/examples/' in text: errors.append(f'{p}: production imports example')
 if 'cmd' in p.parts and '/examples/' in text: errors.append(f'{p}: production imports example')
for p in Path('frontend/src').rglob('*'):
 if p.suffix not in ('.tsx','.mjs'): continue
 text=p.read_text()
 if 'style={' in text or 'style:' in text: errors.append(f'{p}: inline style bypasses shared stylesheet')
for e in errors: print(e)
print('architecture checks: '+('FAIL' if errors else 'PASS'))
sys.exit(bool(errors))
