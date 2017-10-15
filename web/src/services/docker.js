import request from '../utils/request';

export function list() {
  return request('/graphql?query={containers{repo%20version%20time%20ports}}', {
    method: 'get',
    headers: {
      'Content-Type': 'application/graphql',
    },
  });
}

export function changePorts(payload) {
  return request(`/graphql?query=mutation{containers:ports(container:"${payload.selectContainer}",value:"${payload.value}"){repo%20version%20time%20ports}}`, {
    method: 'get',
    headers: {
      'Content-Type': 'application/graphql',
    },
  });
}
