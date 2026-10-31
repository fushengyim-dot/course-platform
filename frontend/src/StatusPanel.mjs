import React, {useEffect, useState} from 'react';
import {classify,messages} from './status.mjs';
const h=React.createElement;
export function StatusView({state,health,onRetry}) {
 return h('section',{'aria-live':'polite','aria-busy':state==='loading'},
  h('h2',null,messages[state]),health&&h(React.Fragment,null,
   h('p',null,'版本：'+health.version),h('ul',null,...Object.entries(health.dependencies).map(([name,status])=>h('li',{key:name},(name==='mysql'?'MySQL':'Qdrant')+'：'+(status==='up'?'正常':'不可用')))),h('p',{className:'muted'},'检查时间：'+health.time)),
  h('button',{disabled:state==='loading',onClick:onRetry},state==='loading'?'检查中…':'重试'));
}
export function StatusPanel(){
 const [state,setState]=useState('loading');const [health,setHealth]=useState(null);
 async function load(){setState('loading');setHealth(null);try{const response=await fetch('/health',{signal:AbortSignal.timeout(4000)});const body=await response.json();const next=classify(response.ok,body);setState(next);if(next==='ready')setHealth(body.data);}catch{setState('error');}}
 useEffect(()=>{void load();},[]);
 return h(StatusView,{state,health,onRetry:()=>void load()});
}
export function StatusPage({details=false,Panel=StatusPanel}){
 return h('main',null,h('header',null,h('h1',null,details?'系统状态详情':'系统状态'),h('p',null,details?'查看版本、检查时间与依赖服务状态。':'查看当前服务是否可以正常使用。'),h('nav',null,h('a',{href:details?'/':'/status/details'},details?'返回系统状态':'查看状态详情'))),h(Panel));
}
