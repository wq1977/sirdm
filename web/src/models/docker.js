import * as dockerservice from '../services/docker';

export default {

  namespace: 'docker',

  state: {
    containers: [],
    portDialogVisible: false,
    portDialogLoading: false,
    logVisible: false,
    inputPorts: '',
    selectContainer: 0,
  },

  subscriptions: {
    setup({ dispatch, history }) {  // eslint-disable-line
      dispatch({ type: 'list' });
    },
  },

  effects: {
    *list({ payload }, { call, put }) {  // eslint-disable-line
      const { data } = yield call(dockerservice.list, { });
      yield put({ type: 'save', payload: { data } });
    },
    *changePorts({ payload }, { call, put, select }) {  // eslint-disable-line
      const docker = yield select(state => state.docker);
      const selectContainerName = docker.containers[docker.selectContainer].repo;
      const { data } = yield call(dockerservice.changePorts, { ...payload,
        selectContainer: selectContainerName });
      yield put({ type: 'save', payload: { data } });
      yield put({ type: 'runtime', payload: { portDialogLoading: false, portDialogVisible: false } });
    },
  },

  reducers: {
    save(state, { payload: { data } }) {
      if (!data) return state;
      const containers = data.data.containers;
      return { ...state, containers };
    },
    runtime(state, { payload }) {
      return { ...state, ...payload };
    },
  },
};
