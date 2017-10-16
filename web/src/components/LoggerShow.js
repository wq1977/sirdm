import React from 'react';
import Websocket from 'react-websocket';
import styles from './LoggerShow.css';

class LoggerShow extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      log: [],
    };
  }
  handleLog(data) {
    this.setState({ log: this.state.log.concat([data]) });
  }
  render() {
    if (!this.props.visible || this.props.repo.length <= 0) {
      return (<div />);
    }
    const logs = this.state.log.map((line) => {
      return (<div>{line}</div>);
    });
    return (
      <div className={styles.normal}>
        Component: LoggerShow <p />
        {logs}
        <Websocket
          url={`ws://192.168.0.148:8088/log?repo=${this.props.repo}`}
          onMessage={this.handleLog.bind(this)}
        />
      </div>
    );
  }
}

export default LoggerShow;
