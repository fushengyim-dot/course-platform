import React, {useEffect, useState} from 'react';
import {createRoot} from 'react-dom/client';
import {classify, messages} from './status.mjs';
import './styles.css';
type State = 'loading'|'ready'|'empty'|'error';
type Health = {version:string;time:string;dependencies:Record<string,string>};
function StatusPanel(){
 const [state,setState]=useState<State>('loading');
 const [health,setHealth]=useState<Health|null>(null);
 async function load(){setState('loading');setHealth(null);try{const response=await fetch('/health',{signal:AbortSignal.timeout(4000)});const body=await response.json();const next=classify(response.ok,body);setState(next);if(next==='ready')setHealth(body.data);}catch{setState('error');}}
 useEffect(()=>{void load();},[]);
 return <main><header><p className="eyebrow">课程项目平台</p><h1>系统状态</h1><p>查看当前服务是否可以正常使用。</p></header><section aria-live="polite" aria-busy={state==='loading'}><h2>{messages[state]}</h2>{health&&<><p>版本：{health.version}</p><ul>{Object.entries(health.dependencies).map(([name,status])=><li key={name}>{name==='mysql'?'MySQL':'Qdrant'}：{status==='up'?'正常':'不可用'}</li>)}</ul><p className="muted">检查时间：{health.time}</p></>}<button disabled={state==='loading'} onClick={()=>void load()}>{state==='loading'?'检查中…':'重试'}</button></section></main>;
}
createRoot(document.getElementById('root')!).render(<StatusPanel/>);
