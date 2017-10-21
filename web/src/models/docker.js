import * as dockerservice from '../services/docker';

export default {

  namespace: 'docker',

  state: {
    containers: [],
    portDialogVisible: false,
    portDialogLoading: false,
    envDialogVisible: false,
    envDialogLoading: false,
    volDialogVisible: false,
    volDialogLoading: false,
    logVisible: false,
    inputPorts: '',
    inputEnv: '',
    inputVol: '',
    selectContainer: 0,
  },

  subscriptions: {
    setup({ dispatch, history }) {  // eslint-disable-line
      dispatch({ type: 'list' });
    },
  },

  effects: {
    *refresh_all({ payload }, { call, put, select }) {
      const docker = yield select(state => state.docker);
      const { data } = yield call(dockerservice.state);
      const containerRunningState = {};
      data.data.state.forEach((e) => {
        containerRunningState[e.repo] = e.state;
      });
      docker.containers.forEach((c, idx) => {
        docker.containers[idx].state = containerRunningState[docker.containers[idx].repo];
      });
      const containers = [].concat(docker.containers);
      yield put({ type: 'runtime', payload: { containers } });
    },
    *list({ payload }, { call, put }) {
      const { data } = yield call(dockerservice.list, { });
      yield put({ type: 'save', payload: { data } });
      yield put({ type: 'refresh_all' });
    },
    *changePorts({ payload }, { call, put, select }) {
      const docker = yield select(state => state.docker);
      const selectContainerName = docker.containers[docker.selectContainer].repo;
      const { data } = yield call(dockerservice.changePorts, { ...payload,
        selectContainer: selectContainerName });
      yield put({ type: 'save', payload: { data } });
      yield put({ type: 'runtime', payload: { portDialogLoading: false, portDialogVisible: false } });
      yield put({ type: 'refresh_all' });
    },
    *changeEnv({ payload }, { call, put, select }) {
      const docker = yield select(state => state.docker);
      const selectContainerName = docker.containers[docker.selectContainer].repo;
      const { data } = yield call(dockerservice.changeEnv, { ...payload,
        selectContainer: selectContainerName });
      yield put({ type: 'save', payload: { data } });
      yield put({ type: 'runtime', payload: { envDialogLoading: false, envDialogVisible: false } });
      yield put({ type: 'refresh_all' });
    },
    *changeVol({ payload }, { call, put, select }) {
      const docker = yield select(state => state.docker);
      const selectContainerName = docker.containers[docker.selectContainer].repo;
      const { data } = yield call(dockerservice.changeVol, { ...payload,
        selectContainer: selectContainerName });
      yield put({ type: 'save', payload: { data } });
      yield put({ type: 'runtime', payload: { volDialogLoading: false, volDialogVisible: false } });
      yield put({ type: 'refresh_all' });
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
