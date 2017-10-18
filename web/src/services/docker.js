import request from '../utils/request';

export function list() {
  return request('/graphql?query={containers{repo%20version%20time%20ports%20env}}', {
    method: 'get',
    headers: {
      'Content-Type': 'application/graphql',
    },
  });
}

export function changePorts(payload) {
  return request(`/graphql?query=mutation{containers:ports(container:"${payload.selectContainer}",value:"${payload.value}"){repo%20version%20time%20ports%20env}}`, {
    method: 'get',
    headers: {
      'Content-Type': 'application/graphql',
    },
  });
}

export function changeEnv(payload) {
  return request(`/graphql?query=mutation{containers:env(container:"${payload.selectContainer}",value:"${payload.value}"){repo%20version%20time%20ports%20env}}`, {
    method: 'get',
    headers: {
      'Content-Type': 'application/graphql',
    },
  });
}

export function state() {
  return request('/graphql?query={state{repo%20state}}', {
    method: 'get',
    headers: {
      'Content-Type': 'application/graphql',
    },
  });
}
