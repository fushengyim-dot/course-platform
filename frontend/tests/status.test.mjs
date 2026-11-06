import {test} from 'node:test';import assert from 'node:assert/strict';
import React from 'react';import {renderToStaticMarkup} from 'react-dom/server';
import {classify,messages} from '../src/status.mjs';import {StatusPage,StatusPanel,StatusView} from '../src/StatusPanel.mjs';
import {InformationPage} from '../src/PublicInformation.mjs';
const data={version:'test-v1',time:'2026-10-31T00:00:00Z',dependencies:{mysql:'up',qdrant:'up'}};
test('S8 normal: successful complete health response',()=>assert.equal(classify(true,{code:'OK',data}),'ready'));
test('S8 failure: invalid structure and service failures',()=>{
 for(const value of [null,{}, {code:'FAIL'}, ...[{},'x',[],{...data,version:''},{...data,time:'bad'},{...data,dependencies:{mysql:'up'}},{...data,dependencies:{...data.dependencies,extra:'up'}},{...data,dependencies:{mysql:'down',qdrant:'up'}}].map(data=>({code:'OK',data}))])assert.equal(classify(true,value),'error');
 assert.equal(classify(false,{code:'OK',data}),'error');
});
test('S8 boundary: empty response',()=>{for(const data of [null,undefined,{dependencies:{}}])assert.equal(classify(true,{code:'OK',data}),'empty')});
test('shared pages reuse the same panel and render every state',()=>{
 for(const details of [false,true]){
  const page=StatusPage({details});assert.equal(page.props.children[1].type,StatusPanel);
  for(const state of Object.keys(messages)){
   const html=renderToStaticMarkup(React.createElement(StatusPage,{details,Panel:()=>React.createElement(StatusView,{state,health:state==='ready'?data:null})}));
   assert.ok(html.includes(messages[state]));assert.ok(html.includes(details?'href="/"':'href="/status/details"'));assert.ok(html.includes('aria-live="polite"'));
   assert.equal(html.includes('disabled=""'),state==='loading');
   if(state==='ready')for(const value of ['test-v1','MySQL','Qdrant',data.time])assert.ok(html.includes(value));
  }
 }
});
test('text interpolation escapes data',()=>assert.ok(renderToStaticMarkup(React.createElement(StatusView,{state:'ready',health:{...data,version:'<script>bad</script>'}})).includes('&lt;script&gt;')));
test('information route renders the institution-library default and preserves direct search',()=>{
 const html=renderToStaticMarkup(React.createElement(InformationPage));
 for(const value of ['信息检索','背景定位','系统填报模拟','个人中心','院校库','直接检索','日历','查找院校','虚构演示数据'])assert.ok(html.includes(value),value);
 assert.ok(!html.includes('来源发布日期起'));
 assert.ok(!html.includes('专业'));
 assert.equal((html.match(/class="portal-nav"/g)||[]).length,1);
});
