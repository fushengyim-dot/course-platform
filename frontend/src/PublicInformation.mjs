import React, {useEffect, useMemo, useState} from 'react';

const h=React.createElement;
const sections=[['information','信息检索'],['positioning','背景定位'],['simulation','系统填报模拟'],['center','个人中心']];
const modes=[['directory','院校库'],['direct','直接检索'],['calendar','日历']];
const emptyFilters={institution:'',college:'',expected_publish_from:'',expected_publish_to:''};

async function readData(url){
 const response=await fetch(url,{signal:AbortSignal.timeout(5000)});
 const body=await response.json();
 if(!response.ok||body.code!=='OK')throw new Error('request failed');
 return body.data;
}

export function InformationPage(){
 const [section,setSection]=useState('information');
 const [mode,setMode]=useState('directory');
 const [institutions,setInstitutions]=useState([]);
 const [directoryState,setDirectoryState]=useState('loading');
 const [directoryRetry,setDirectoryRetry]=useState(0);
 const [directorySearch,setDirectorySearch]=useState('');
 const [institutionDetail,setInstitutionDetail]=useState(null);
 const [filters,setFilters]=useState({...emptyFilters});
 const [applied,setApplied]=useState({...emptyFilters});
 const [page,setPage]=useState(null);
 const [selectedInfo,setSelectedInfo]=useState(null);
 const [searchState,setSearchState]=useState('loading');
 const [searchRetry,setSearchRetry]=useState(0);
 const [calendarMonth,setCalendarMonth]=useState(()=>{const now=new Date();return `${now.getFullYear()}-${String(now.getMonth()+1).padStart(2,'0')}`;});
 const [calendarData,setCalendarData]=useState(null);
 const [calendarState,setCalendarState]=useState('loading');
 const [calendarRetry,setCalendarRetry]=useState(0);
 const [selectedCalendarDay,setSelectedCalendarDay]=useState('');

 useEffect(()=>{
  let active=true;
  setDirectoryState('loading');
  readData('/api/v1/public-institutions').then(items=>{
   if(active){setInstitutions(items);setDirectoryState(items.length?'ready':'empty');}
  }).catch(()=>{if(active)setDirectoryState('error');});
  return()=>{active=false;};
 },[directoryRetry]);

 useEffect(()=>{
  if(section!=='information'||mode!=='direct'||selectedInfo)return;
  let active=true;
  setSearchState('loading');
  const params=new URLSearchParams();
  for(const [key,value] of Object.entries(applied))if(value)params.set(key,value);
  readData('/api/v1/public-information?'+params).then(result=>{
   if(active){setPage(result);setSearchState(result.items.length?'ready':'empty');}
  }).catch(()=>{if(active)setSearchState('error');});
  return()=>{active=false;};
 },[section,mode,selectedInfo,applied,searchRetry]);

 useEffect(()=>{
  if(section!=='information'||mode!=='calendar')return;
  let active=true;
  setCalendarState('loading');
  readData('/api/v1/public-information/calendar?month='+encodeURIComponent(calendarMonth)).then(result=>{
   if(active){setCalendarData(result);setCalendarState('ready');setSelectedCalendarDay(previous=>previous.startsWith(calendarMonth)?previous:(result.days[0]?.date||calendarMonth+'-01'));}
  }).catch(()=>{if(active)setCalendarState('error');});
  return()=>{active=false;};
 },[section,mode,calendarMonth,calendarRetry]);

 const selectedInstitution=useMemo(()=>institutions.find(item=>item.name===filters.institution||item.aliases?.some(alias=>alias===filters.institution))||null,[institutions,filters.institution]);
 const visibleInstitutions=useMemo(()=>{
  const term=directorySearch.trim().toLocaleLowerCase();
  if(!term)return institutions;
  return institutions.filter(item=>[item.name,...(item.aliases||[])].some(name=>name.toLocaleLowerCase().includes(term)));
 },[institutions,directorySearch]);
 const update=(key,value)=>setFilters(previous=>({...previous,[key]:value}));

 async function openInstitution(item){
  setInstitutionDetail({loading:true});
  try{setInstitutionDetail(await readData('/api/v1/public-institutions/'+encodeURIComponent(item.id)));}
  catch{setInstitutionDetail({error:true});}
 }
 async function openInformation(item){
  setSelectedInfo({loading:true});
  try{setSelectedInfo(await readData('/api/v1/public-information/'+encodeURIComponent(item.id)));}
  catch{setSelectedInfo({error:true});}
 }
 async function loadMore(){
  if(!page?.next_cursor)return;
  const params=new URLSearchParams();
  for(const [key,value] of Object.entries(applied))if(value)params.set(key,value);
  params.set('cursor',page.next_cursor);
  try{const next=await readData('/api/v1/public-information?'+params);setPage({...next,items:[...page.items,...next.items]});}
  catch{setSearchState('error');}
 }

 const sectionNav=h('nav',{className:'portal-nav','aria-label':'主栏目'},...sections.map(([id,label])=>h('button',{key:id,type:'button',className:section===id?'active':'',onClick:()=>{setSection(id);setInstitutionDetail(null);setSelectedInfo(null);},'aria-current':section===id?'page':undefined},label)));
 const header=h('header',{className:'portal-header'},
  h('div',null,h('p',{className:'portal-kicker'},'学生服务平台'),h('h1',null,'信息服务'),h('p',null,'按院校浏览招生信息，也可以直接筛选。')),
  h('nav',{className:'utility-nav'},h('a',{href:'/'},'系统状态'),h('a',{href:'/status/details'},'状态详情'),h('button',{type:'button',onClick:()=>window.alert('登录与注册将在后续版本接入。')},'登录 / 注册')));

 let content;
 if(section!=='information'){
  content=h('section',{className:'plain-panel'},h('p',null,section==='center'?'个人资料和个人记录将在后续版本提供。':'此栏目将在后续版本提供。'));
 }else{
  const modeNav=h('nav',{className:'mode-nav','aria-label':'信息检索方式'},...modes.map(([id,label])=>h('button',{key:id,type:'button',className:mode===id?'active':'',onClick:()=>{setMode(id);setInstitutionDetail(null);setSelectedInfo(null);},'aria-current':mode===id?'page':undefined},label)));
  if(mode==='calendar'){
   const [year,monthNumber]=calendarMonth.split('-').map(Number);
   const monthStart=new Date(year,monthNumber-1,1);
   const daysInMonth=new Date(year,monthNumber,0).getDate();
   const mondayOffset=(monthStart.getDay()+6)%7;
   const dayMap=new Map((calendarData?.days||[]).map(day=>[day.date,day.items]));
   const monthLabel=`${year}年${String(monthNumber).padStart(2,'0')}月`;
   const moveMonth=delta=>{const date=new Date(year,monthNumber-1+delta,1);setCalendarMonth(`${date.getFullYear()}-${String(date.getMonth()+1).padStart(2,'0')}`);setSelectedCalendarDay('');};
   const cellCount=Math.ceil((mondayOffset+daysInMonth)/7)*7;
   const cells=Array.from({length:cellCount},(_,index)=>{
    if(index<mondayOffset||index>=mondayOffset+daysInMonth)return h('span',{key:'blank-'+index,className:'calendar-blank','aria-hidden':'true'});
    const dayNumber=index-mondayOffset+1;
    const date=`${calendarMonth}-${String(dayNumber).padStart(2,'0')}`;
    const count=dayMap.get(date)?.length||0;
    return h('button',{key:date,type:'button',className:`calendar-day${count?' has-items':''}${selectedCalendarDay===date?' selected':''}`,onClick:()=>setSelectedCalendarDay(date),'aria-pressed':selectedCalendarDay===date,'aria-label':`${date}，${count}条预计发布信息`},h('span',null,String(dayNumber)),count>0&&h('small',null,`${count}条`));
   });
   const entries=dayMap.get(selectedCalendarDay)||[];
   content=h(React.Fragment,null,
    modeNav,
    h('section',{className:'calendar-panel'},
     h('div',{className:'calendar-toolbar'},
      h('button',{type:'button',className:'secondary-button',onClick:()=>moveMonth(-1),'aria-label':'上一个月'},'‹ 上个月'),
      h('h3',null,monthLabel),
      h('button',{type:'button',className:'secondary-button',onClick:()=>moveMonth(1),'aria-label':'下一个月'},'下个月 ›')),
     h('p',{className:'data-note'},'按预计发布时间查看招生信息；虚构演示日期不代表学校实际发布计划。'),
     calendarState==='loading'
      ? h('p',{role:'status'},'正在读取日历…')
      : calendarState==='error'
       ? h('div',{role:'alert'},h('p',null,'日历读取失败。'),h('button',{type:'button',onClick:()=>setCalendarRetry(value=>value+1)},'重试'))
       : h(React.Fragment,null,
        h('div',{className:'calendar-grid','aria-label':monthLabel},
         ...['一','二','三','四','五','六','日'].map((label,index)=>h('span',{key:'weekday-'+index,className:'calendar-weekday'},label)),
         ...cells),
        h('section',{className:'calendar-day-results','aria-live':'polite'},
         h('h3',null,selectedCalendarDay+' 的预计发布信息'),
         entries.length
          ? h('div',{className:'results'},...entries.map(entry=>h('article',{key:entry.id,className:'result-row'},
             h('h4',null,entry.institution_name),
             h('p',{className:'meta'},(entry.college_names||[]).join(' · ')),
             h('p',null,entry.title),
             h('a',{href:entry.registration_url,target:'_blank',rel:'noreferrer'},'报名入口（虚构演示链接）'))))
          : h('p',null,'这一天没有预计发布信息。')))));
  }else if(mode==='directory'&&institutionDetail){
   const detailContent=institutionDetail.loading?h('p',{role:'status'},'正在读取院校资料…'):institutionDetail.error?h('div',{role:'alert'},h('p',null,'院校资料读取失败。'),h('button',{type:'button',onClick:()=>setInstitutionDetail(null)},'返回院校库')):h(React.Fragment,null,
    h('p',{className:'meta'},institutionDetail.region),h('p',null,institutionDetail.description),
    h('h3',null,'学院'),h('ul',{className:'plain-list'},...(institutionDetail.colleges||[]).map(college=>h('li',{key:college},college))),
    h('h3',null,'招生信息与报名入口'),h('p',{className:'data-note'},'链接为虚构演示地址，不能用于真实报名。'),
    h('div',{className:'results'},...(institutionDetail.admission_entries||[]).map(entry=>h('article',{className:'result-row',key:entry.id},
     h('h4',null,entry.title),h('p',{className:'meta'},'预计发布时间：'+(entry.expected_publication_on||'暂未估计')),
     h('a',{href:entry.registration_url,target:'_blank',rel:'noreferrer'},'报名入口（演示链接）'),
     h('p',{className:'meta'},'来源：'+entry.source_title+'；来源标注日期：'+(entry.source_published_on||'未提供'))))));
   content=h(React.Fragment,null,modeNav,h('section',{className:'plain-panel'},h('button',{type:'button',className:'text-button',onClick:()=>setInstitutionDetail(null)},'← 返回院校库'),h('h2',null,institutionDetail.name||'院校资料'),detailContent));
  }else if(mode==='directory'){
   content=h(React.Fragment,null,modeNav,h('section',{className:'directory-panel'},
    h('label',{className:'directory-search'},'查找院校',h('input',{type:'search',value:directorySearch,onChange:e=>setDirectorySearch(e.target.value),placeholder:'输入学校名称或简称'})),
    h('p',{className:'data-note'},'院校与报名信息均为虚构演示数据。'),
    directoryState==='loading'?h('p',{role:'status'},'正在读取院校库…'):directoryState==='error'?h('div',{role:'alert'},h('p',null,'院校库暂时不可用。'),h('button',{type:'button',onClick:()=>setDirectoryRetry(value=>value+1)},'重试')):visibleInstitutions.length===0?h('p',null,'没有找到院校。'):h('div',{className:'institution-list'},...visibleInstitutions.map(item=>h('article',{className:'institution-row',key:item.id},
     h('button',{type:'button',className:'institution-link',onClick:()=>openInstitution(item)},item.name),h('p',{className:'meta'},item.region+'　·　'+item.colleges.length+' 个院系'),h('p',null,item.description))))));
  }else if(selectedInfo){
   const detailContent=selectedInfo.loading?h('p',{role:'status'},'正在读取…'):selectedInfo.error?h('p',{role:'alert'},'详情读取失败，请返回后重试。'):h(React.Fragment,null,
    h('h2',null,selectedInfo.title),h('p',{className:'meta'},[selectedInfo.institution_name,...(selectedInfo.college_names||[])].join(' · ')),h('p',null,selectedInfo.body),
    h('dl',{className:'detail-meta'},h('dt',null,'预计发布时间'),h('dd',null,selectedInfo.expected_publication_on||'暂未估计'),h('dt',null,'来源'),h('dd',null,selectedInfo.source_title),h('dt',null,'来源标注日期'),h('dd',null,selectedInfo.source_published_on||'未提供'),h('dt',null,'报名入口'),h('dd',null,h('a',{href:selectedInfo.registration_url,target:'_blank',rel:'noreferrer'},'报名入口（虚构演示链接）')),
     h('dt',null,'数据说明'),h('dd',null,'预计日期和报名链接均为虚构演示数据，无真实来源')));
   content=h(React.Fragment,null,modeNav,h('section',{className:'plain-panel'},h('button',{type:'button',className:'text-button',onClick:()=>setSelectedInfo(null)},'← 返回列表'),detailContent));
  }else{
   const collegeOptions=selectedInstitution?.colleges||[];
   const form=h('form',{className:'search-form',onSubmit:e=>{e.preventDefault();setApplied({...filters});}},
    h('div',{className:'field-grid'},
     h('label',null,'学校',h('input',{list:'institution-suggestions',value:filters.institution,onChange:e=>setFilters(previous=>({...previous,institution:e.target.value,college:''})),maxLength:80,autoComplete:'off'}),
      h('datalist',{id:'institution-suggestions'},...institutions.flatMap(item=>[item.name,...(item.aliases||[])].map(name=>h('option',{key:item.id+'-'+name,value:name}))))),
     h('label',null,'学院',h('select',{value:filters.college,disabled:!selectedInstitution,onChange:e=>update('college',e.target.value)},h('option',{value:''},selectedInstitution?'全部学院':'请先选择学校'),...collegeOptions.map(college=>h('option',{key:college,value:college},college)))),
     h('label',null,'预计发布时间起',h('input',{type:'date',value:filters.expected_publish_from,onChange:e=>update('expected_publish_from',e.target.value)})),
     h('label',null,'预计发布时间止',h('input',{type:'date',value:filters.expected_publish_to,onChange:e=>update('expected_publish_to',e.target.value)}))),
    h('div',{className:'form-actions'},h('button',{type:'submit'},'检索'),h('button',{type:'button',className:'secondary-button',onClick:()=>{setFilters({...emptyFilters});setApplied({...emptyFilters});}},'清空')));
   const results=searchState==='loading'?h('p',{role:'status'},'正在检索…'):searchState==='error'?h('div',{role:'alert'},h('p',null,'检索暂时失败。'),h('button',{type:'button',onClick:()=>setSearchRetry(value=>value+1)},'重试')):searchState==='empty'?h('p',null,'没有找到符合条件的条目。'):h('div',{className:'results','aria-live':'polite'},...(page?.items||[]).map(({information:item,matched_terms:terms})=>h('article',{className:'result-row',key:item.id},
    h('button',{type:'button',className:'result-title',onClick:()=>openInformation(item)},item.title),h('p',{className:'meta'},[item.institution_name,...(item.college_names||[])].join(' · ')),h('p',null,item.body),h('p',{className:'meta'},'预计发布时间：'+(item.expected_publication_on||'暂未估计')+'　·　匹配：'+(terms||[]).join('、')))));
   content=h(React.Fragment,null,modeNav,form,h('p',{className:'data-note'},'按预计发布时间范围筛选招生信息；预计日期可能与实际发布时间不同。院校和报名入口均为虚构演示数据。'),results,page?.next_cursor&&h('button',{type:'button',className:'secondary-button',onClick:loadMore},'加载更多'));
  }
 }
 return h('main',{className:'portal'},header,sectionNav,h('div',{className:'section-heading'},h('h2',null,sections.find(([id])=>id===section)?.[1])),content,h('footer',null,'演示版本 · 数据为虚构样例'));
}
