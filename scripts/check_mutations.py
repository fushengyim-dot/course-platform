from pathlib import Path
import subprocess

def rejected(name,path,mutate,command,cwd=None):
 p=Path(path);original=p.read_bytes()
 try:
  changed=mutate(original.decode())
  if changed==original.decode(): raise RuntimeError(name+": mutation did not apply")
  p.write_text(changed)
  r=subprocess.run(command,cwd=cwd,capture_output=True,text=True)
  if r.returncode==0: raise RuntimeError(name+": violation was not detected")
  print(name+": correctly rejected")
 finally: p.write_bytes(original)
rejected("reverse dependency","backend/internal/service/health.go",lambda s:s.replace('import (','import (\n _ \"group-platform/internal/handler\"',1),["python3","scripts/check_architecture.py"])
rejected("inline style","frontend/src/main.tsx",lambda s:s.replace('<main>','<main style={{margin:13}}>'),["python3","scripts/check_architecture.py"])
rejected("removed authorization","backend/internal/handler/http.go",lambda s:s.replace('status = 401','status = 200'),["go","test","./internal/handler","-run","TestFailureRoutes","-count=1"],"backend")
