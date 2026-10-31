export function classify(responseOK, body) {
 if (!responseOK || body?.code !== 'OK') return 'error';
 if (!body.data || !body.data.dependencies || Object.keys(body.data.dependencies).length === 0) return 'empty';
 if (Object.values(body.data.dependencies).some(value => value !== 'up')) return 'error';
 return 'ready';
}
export const messages = {loading:'正在检查服务状态…',ready:'服务已就绪',empty:'暂无状态数据，请重试',error:'服务暂不可用，请稍后重试；持续失败请检查运行环境'};
