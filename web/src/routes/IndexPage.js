import React from 'react';
import { connect } from 'dva';
import Button from 'antd/lib/button';
import 'antd/lib/button/style';
import Row from 'antd/lib/row';
import 'antd/lib/row/style';
import Col from 'antd/lib/col';
import 'antd/lib/col/style';
import Modal from 'antd/lib/modal';
import 'antd/lib/modal/style';
import Input from 'antd/lib/input';
import 'antd/lib/input/style';
import styles from './IndexPage.css';
import LoggerShow from '../components/LoggerShow';

let input;
let envinput;
function handleChangePort(props) {
  props.dispatch({ type: 'docker/runtime', payload: { portDialogLoading: true } });
  props.dispatch({ type: 'docker/changePorts', payload: { value: input.refs.input.value } });
}

function handleChangeEnv(props) {
  props.dispatch({ type: 'docker/runtime', payload: { envDialogLoading: true } });
  props.dispatch({ type: 'docker/changeEnv', payload: { value: envinput.refs.input.value } });
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

function showPortDialog(props, idx) {
  props.dispatch({
    type: 'docker/runtime',
    payload: { inputPorts: props.docker.containers[idx].ports } });
  props.dispatch({
    type: 'docker/runtime',
    payload: { portDialogVisible: true, selectContainer: idx } });
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
  const contains = [''].concat(props.docker.containers).map((value, index) => {
    if (index === 0) {
      return (
        <Row type="flex" key={index}>
          <Col span={3}>名称</Col>
          <Col span={3}>启动时间</Col>
          <Col span={3}>运行状态</Col>
          <Col span={3}>运行镜像版本</Col>
          <Col span={3}>开放端口</Col>
          <Col span={6}>环境变量</Col>
          <Col span={3}>操作</Col>
        </Row>
      );
    }
    const envs = value.env.split('|').map((env, idx) => {
      return (<Row key={idx}>{env}</Row>);
    });
    return (
      <Row type="flex" align="middle" key={index} style={{ marginTop: 10 }}>
        <Col type="flex" span={3}>{value.repo}</Col>
        <Col span={3}>{value.time}</Col>
        <Col span={3}>{value.state}</Col>
        <Col span={3}>{value.version.substring(7, 23)}</Col>
        <Col span={3}>{value.ports}</Col>
        <Col span={6}>{envs}</Col>
        <Col span={3} type="flex">
          <Row>
            <Col>
              <Button
                onClick={showLog.bind(null, props, index - 1)}
                className={styles.button}
              >查看日志</Button>
            </Col>
            <Col>
              <Button
                onClick={showPortDialog.bind(null, props, index - 1)}
                className={styles.button}
              >修改端口</Button>
            </Col>
            <Col>
              <Button
                onClick={showEnvDialog.bind(null, props, index - 1)}
                className={styles.button}
              >环境变量</Button>
            </Col>
          </Row>
        </Col>
      </Row>
    );
  });
  const { envDialogVisible, envDialogLoading,
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
        <div className={styles.list}>
          {contains}
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
