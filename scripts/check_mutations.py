from pathlib import Path
import subprocess

def rejected(name,path,mutate,command,cwd=None):
 p=Path(path);original=p.read_bytes()
 try:
  changed=mutate(original.decode())
  if changed==original.decode(): raise RuntimeError(name+": mutation did not apply")
  p.write_text(changed)
  r=subprocess.run(command,cwd=cwd,capture_output=True,text=True,timeout=120)
  if r.returncode==0: raise RuntimeError(name+": violation was not detected")
  if command[0]=="go" and "--- FAIL:" not in r.stdout+r.stderr: raise RuntimeError(name+": failed without a test assertion: "+r.stdout+r.stderr)
  print(name+": correctly rejected")
 finally: p.write_bytes(original)
rejected("reverse dependency","backend/internal/service/health.go",lambda s:s.replace('import (','import (\n _ \"group-platform/internal/handler\"',1),["python3","scripts/check_architecture.py"])
rejected("inline style","frontend/src/StatusPanel.mjs",lambda s:s.replace("h('main',null", "h('main',{style:{margin:13}}"),["python3","scripts/check_architecture.py"])
rejected("removed authorization","backend/internal/handler/http.go",lambda s:s.replace('status = 401','status = 200'),["go","test","./internal/handler","-run","TestFailureRoutes","-count=1"],"backend")

rejected("example reverse dependency","backend/examples/records/service/records.go",lambda s:s.replace('import (','import (\n _ "group-platform/examples/records/handler"',1),["python3","scripts/check_architecture.py"])
rejected("example removed identity","backend/examples/records/handler/http.go",lambda s:s.replace('if !ok || strings.TrimSpace(subject) == ""','if (!ok || strings.TrimSpace(subject) == "") && false'),["go","test","./examples/records/handler","-run","TestFailureUnauthorized","-count=1"],"backend")
rejected("example removed ownership","backend/examples/records/service/records.go",lambda s:s.replace('r.Subject != subject','false'),["go","test","./examples/records/handler","-run","TestFailureOwnership","-count=1"],"backend")
