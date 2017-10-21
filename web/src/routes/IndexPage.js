import React from 'react';
import { connect } from 'dva';
import Button from 'antd/lib/button';
import 'antd/lib/button/style';
import Row from 'antd/lib/row';
import 'antd/lib/row/style';
import Table from 'antd/lib/table';
import 'antd/lib/table/style';
import Col from 'antd/lib/col';
import 'antd/lib/col/style';
import Modal from 'antd/lib/modal';
import 'antd/lib/modal/style';
import Input from 'antd/lib/input';
import 'antd/lib/input/style';
import moment from 'moment';
import styles from './IndexPage.css';
import LoggerShow from '../components/LoggerShow';

let input;
let envinput;
let volinput;
function handleChangePort(props) {
  props.dispatch({ type: 'docker/runtime', payload: { portDialogLoading: true } });
  props.dispatch({ type: 'docker/changePorts', payload: { value: input.refs.input.value } });
}

function handleChangeEnv(props) {
  props.dispatch({ type: 'docker/runtime', payload: { envDialogLoading: true } });
  props.dispatch({ type: 'docker/changeEnv', payload: { value: envinput.refs.input.value } });
}

function handleChangeVol(props) {
  props.dispatch({ type: 'docker/runtime', payload: { volDialogLoading: true } });
  props.dispatch({ type: 'docker/changeVol', payload: { value: volinput.refs.input.value } });
}

function showLog(props, idx) {
  props.dispatch({
    type: 'docker/runtime',
    payload: { logVisible: true, selectContainer: idx } });
}

function showEnvDialog(props, idx) {
  props.dispatch({
    type: 'docker/runtime',
    payload: { inputEnv: props.docker.containers[idx].env } });
  props.dispatch({
    type: 'docker/runtime',
    payload: { envDialogVisible: true, selectContainer: idx } });
}

function showVolDialog(props, idx) {
  props.dispatch({
    type: 'docker/runtime',
    payload: {
      inputVols: props.docker.containers[idx].vols,
      volDialogVisible: true,
      selectContainer: idx,
    },
  });
}

function showPortDialog(props, idx) {
  props.dispatch({
    type: 'docker/runtime',
    payload: { inputPorts: props.docker.containers[idx].ports } });
  props.dispatch({
    type: 'docker/runtime',
    payload: { portDialogVisible: true, selectContainer: idx } });
}

function handleVolCancel(props) {
  props.dispatch({ type: 'docker/runtime', payload: { volDialogVisible: false } });
}

function handleLogCancel(props) {
  props.dispatch({ type: 'docker/runtime', payload: { logVisible: false } });
}

function handleEnvCancel(props) {
  props.dispatch({ type: 'docker/runtime', payload: { envDialogVisible: false } });
}

function handleCancel(props) {
  props.dispatch({ type: 'docker/runtime', payload: { portDialogVisible: false } });
}

function IndexPage(props) {
  const dataSource = props.docker.containers.map((value, idx) => {
    const v = { ...value };
    v.version = value.version.substring(7, 23);
    const envs = value.env.split('|').map((env, i) => {
      return (<Row key={i}>{env}</Row>);
    });
    v.envs = envs;
    v.time = moment(v.time).format('YYYY-MM-DD HH:mm:ss');
    const btns = (<Row>
      <Row>
        <Col>
          <Button
            onClick={showLog.bind(null, props, idx)}
            className={styles.button}
          >查看日志</Button>
        </Col>
        <Col>
          <Button
            onClick={showPortDialog.bind(null, props, idx)}
            className={styles.button}
          >修改端口</Button>
        </Col>
      </Row>
      <Row>
        <Col>
          <Button
            onClick={showEnvDialog.bind(null, props, idx)}
            className={styles.button}
          >环境变量</Button>
        </Col>
        <Col>
          <Button
            onClick={showVolDialog.bind(null, props, idx)}
            className={styles.button}
          >挂载点</Button>
        </Col>
      </Row>
    </Row>);
    v.btns = btns;
    return v;
  });
  const columns = [{
    title: '名称',
    dataIndex: 'repo',
    key: 'repo',
  }, {
    title: '启动时间',
    dataIndex: 'time',
    key: 'time',
  }, {
    title: '运行状态',
    dataIndex: 'state',
    key: 'state',
  }, {
    title: '运行镜像版本',
    dataIndex: 'version',
    key: 'version',
  }, {
    title: '开放端口',
    dataIndex: 'ports',
    key: 'ports',
  }, {
    title: '环境变量',
    dataIndex: 'envs',
    key: 'envs',
  }, {
    title: '挂载',
    dataIndex: 'vols',
    key: 'vols',
  }, {
    title: '操作',
    dataIndex: 'btns',
    key: 'btns',
    className: styles.btns,
  }];
  const { volDialogVisible, volDialogLoading, envDialogVisible, envDialogLoading,
    portDialogVisible, portDialogLoading, logVisible } = props.docker;
  let selectRepo = '';
  if (props.docker.selectContainer < props.docker.containers.length) {
    selectRepo = props.docker.containers[props.docker.selectContainer].repo;
  }
  if (!logVisible) {
    return (
      <div className={styles.normal}>
        <h1 className={styles.title}>SIRDM</h1>
        <h2>运行中的镜像</h2>
        <div className={styles.table}>
          <Table dataSource={dataSource} rowKey="repo" columns={columns} pagination={false} />
        </div>
        <Modal
          visible={portDialogVisible}
          title="修改容器端口"
          onOk={handleChangePort.bind(null, props)}
          onCancel={handleCancel.bind(null, props)}
          footer={[
            <Button key="back" size="large" onClick={handleCancel.bind(null, props)}>取消</Button>,
            <Button key="submit" type="primary" size="large" loading={portDialogLoading} onClick={handleChangePort.bind(null, props)}>
              确定
            </Button>,
          ]}
        >
          <Input
            defaultValue={props.docker.inputPorts}
            ref={c => (input = c)}
            placeholder="请输入要暴漏的端口，逗号分割"
          />
        </Modal>
        <Modal
          visible={envDialogVisible}
          title="修改容器环境变量"
          onOk={handleChangeEnv.bind(null, props)}
          onCancel={handleEnvCancel.bind(null, props)}
          footer={[
            <Button key="back" size="large" onClick={handleEnvCancel.bind(null, props)}>取消</Button>,
            <Button key="submit" type="primary" size="large" loading={envDialogLoading} onClick={handleChangeEnv.bind(null, props)}>
              确定
            </Button>,
          ]}
        >
          <Input
            defaultValue={props.docker.inputEnv}
            ref={c => (envinput = c)}
            placeholder="请输入要设置的环境变量，|分割"
          />
        </Modal>
        <Modal
          visible={volDialogVisible}
          title="修改容器挂载点"
          onOk={handleChangeVol.bind(null, props)}
          onCancel={handleVolCancel.bind(null, props)}
          footer={[
            <Button key="back" size="large" onClick={handleVolCancel.bind(null, props)}>取消</Button>,
            <Button key="submit" type="primary" size="large" loading={volDialogLoading} onClick={handleChangeVol.bind(null, props)}>
              确定
            </Button>,
          ]}
        >
          <Input
            defaultValue={props.docker.inputVol}
            ref={c => (volinput = c)}
            placeholder="请输入要设置的环境变量，|分割"
          />
        </Modal>
      </div>
    );
  }
  return (
    <div>
      <Button
        onClick={handleLogCancel.bind(null, props)}
        className={styles.closeLog}
      >关闭</Button>
      <LoggerShow visible={logVisible} repo={selectRepo} />
    </div>
  );
}

IndexPage.propTypes = {
};

export default connect(
  ({ docker }) => ({ docker }),
)(IndexPage);
