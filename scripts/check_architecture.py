from pathlib import Path
import re
import sys
errors=[]
for p in Path('backend/internal').rglob('*.go'):
 if p.name.endswith('_test.go'): continue
 layer=p.parts[2]
 text=p.read_text()
 forbidden={'service':['net/http','database/sql','internal/adapter','internal/handler'], 'repository':['internal/service','internal/handler','internal/adapter'], 'handler':['database/sql','internal/adapter','go-sql-driver']}.get(layer,[])
 for value in forbidden:
  if value in text: errors.append(f'{p}: forbidden dependency {value}')
 if layer=='service' and re.search(r'(?i)\b(select|insert|update|delete)\s+.*\b(from|into|set)\b',text): errors.append(f'{p}: SQL in service')
for p in Path('frontend/src').glob('*.tsx'):
 if 'style={' in p.read_text(): errors.append(f'{p}: inline style bypasses shared stylesheet')
for e in errors: print(e)
print('architecture checks: '+('FAIL' if errors else 'PASS'))
sys.exit(bool(errors))
