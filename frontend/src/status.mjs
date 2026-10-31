export function classify(responseOK,body){
 if(!responseOK||body?.code!=='OK')return 'error';
 const d=body.data;
 if(d==null)return 'empty';
 if(typeof d!=='object'||Array.isArray(d))return 'error';
 const deps=d.dependencies;
 if(deps&&typeof deps==='object'&&!Array.isArray(deps)&&Object.keys(deps).length===0)return 'empty';
 if(typeof d.version!=='string'||!d.version.trim()||typeof d.time!=='string'||!d.time.endsWith('Z')||!Number.isFinite(Date.parse(d.time)))return 'error';
 if(!deps||Array.isArray(deps)||typeof deps!=='object'||Object.keys(deps).length!==2||deps.mysql!=='up'||deps.qdrant!=='up')return 'error';
 return 'ready';
}
export const messages={loading:'正在检查服务状态…',ready:'服务已就绪',empty:'暂无状态数据，请重试',error:'服务暂不可用，请稍后重试；持续失败请检查运行环境'};
