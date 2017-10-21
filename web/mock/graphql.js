module.exports = {
  'GET /graphql': (req, res) => {
    if (req.url.indexOf('{containers{') >= 0) {
      res.json({
        data: {
          containers: [
            {
              env: 'MYSQL_ROOT_PASSWORD=my-secret-pw',
              ports: '9883,9882',
              time: '2017-10-17T00:00:19+08:00',
              version: 'sha256:2fcbd08f3740198e15fb7437f85ad1132be27a210877e8ed327deb0d52c2f3eb',
              repo: 'gateway',
              vols: '/var/log/freego:/var/log/freego',
            },
            {
              env: 'MYSQL_ROOT_PASSWORD=my-secret-pw|MYSQL_ROOT_PASSWORD=my-secret-p',
              ports: '9883,9882',
              time: '2017-10-17T00:00:19+08:00',
              version: 'sha256:2fcbd08f3740198e15fb7437f85ad1132be27a210877e8ed327deb0d52c2f3eb',
              repo: 'freego',
              vols: '/var/log/freego:/var/log/freego',
            },
          ],
        },
      });
      return;
    }
    if (req.url.indexOf('{state{') >= 0) {
      res.json({
        data: {
          state: [
            {
              repo: 'gateway',
              state: 'running',
            },
            {
              repo: 'freego',
              state: 'exting',
            },
          ],
        },
      });
      return;
    }
    res.json({
      success: true,
    });
  },
};
