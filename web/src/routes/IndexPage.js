import React from 'react';
import { connect } from 'dva';
import Button from 'antd/lib/button';
import 'antd/lib/button/style';
import Row from 'antd/lib/row';
import 'antd/lib/row/style';
import Col from 'antd/lib/col';
import 'antd/lib/col/style';
import styles from './IndexPage.css';

function IndexPage() {
  const contains = ['', 'freego'].map((value, index) => {
    if (index === 0) {
      return (
        <Row type="flex" justify="center" key={index}>
          <Col span={3} className={styles.col1}>名称</Col>
          <Col span={9} className={styles.col2}>当前状态</Col>
          <Col span={6} className={styles.col3}>运行版本</Col>
          <Col span={6} className={styles.col4}>操作</Col>
        </Row>
      );
    }
    return (
      <Row type="flex" justify="center" key={index}>
        <Col span={3} className={styles.col1}>{value}</Col>
        <Col span={9} className={styles.col2}>当前状态</Col>
        <Col span={6} className={styles.col3}>运行版本</Col>
        <Col span={6} className={styles.col4}>
          <Button className={styles.button}>查看日志</Button>
        </Col>
      </Row>
    );
  });
  return (
    <div className={styles.normal}>
      <h1 className={styles.title}>SIRDM 也许不是一个好名字，但是这不重要!</h1>
      <h2>运行中的镜像</h2>
      <div className={styles.list}>
        {contains}
      </div>
    </div>
  );
}

IndexPage.propTypes = {
};

export default connect()(IndexPage);
