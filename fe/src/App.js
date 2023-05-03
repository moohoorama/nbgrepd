import 'antd/dist/antd.css';

import React  from 'react';

import { Dropdown, Layout, Menu, message, Space } from 'antd';
import { DownOutlined } from '@ant-design/icons';
import { useHistory, useLocation } from 'react-router';
import { Link, Switch, Route } from 'react-router-dom';
import { useEffect, useState } from 'react';

import { StatusList } from './Status.js';
import axios from 'axios';

const { SubMenu } = Menu;
const { Header, Content, Sider } = Layout;

function App() {
  return <Main />
}


function Main() {
  let location = useLocation();
  const [selectedMenu, setSelectedMenu] = useState([""]);
  const [apiURL, setAPIURL] = useState("");


  useEffect(() => {
    let firstPath = location.pathname.split("/")[1]
    setSelectedMenu([firstPath])
  }, [location]);

  useEffect(() => { axios.get("/api_url").then(res => setAPIURL(res.data.apiurl)) }, []);

  let history = useHistory();

  const requestLogout = () => {
    axios.get("/auth/logout")
      .then(() => history.go(0))
      .catch(e => message.error(e.toString()))
  }

  const accountMenu = <Menu>
      <Menu.Item>
        <a href='#!' onClick={() => requestLogout() }>Logout</a>
      </Menu.Item>
    </Menu>

  return (
    <>
    <Header>
      <Link to="/" style={{ fontSize: "18px" }} onClick={() => setSelectedMenu([])}>TSCoke Admin</Link>
      <Space style={{ float: "right" }}>
        <Dropdown overlay={accountMenu}>
          <a href='#!'> testtest <DownOutlined /> </a>
        </Dropdown>
      </Space>
    </Header>
    <Layout>
      <Sider width={200} className="site-layout-background">
        <Menu
          mode="inline"
          selectedKeys={selectedMenu}
          defaultOpenKeys={['sub1', 'sub2', 'sub3']}
          style={{ height: '100%', borderRight: 0 }}
        >
          <Menu.Item key="" title="Dashboard" onClick={() => history.push("/")} >Dashboard</Menu.Item>
          <Menu.Item key="account" title="account" onClick={() => history.push("/account")} >Account</Menu.Item>
          <SubMenu key="sub1" title="TSD" >
            <Menu.Item key="tsd" title="TSD" onClick={() => history.push("/tsd")}>List</Menu.Item>
            <Menu.Item key="tsd_env" title="Env" onClick={() => history.push("/tsd_env")}>Env</Menu.Item>
            <Menu.Item key="tsd_tickets" title="Tickets" onClick={() => history.push("/tsd_tickets")}>Tickets</Menu.Item>
          </SubMenu>
          
          <SubMenu key="sub2" title="QD">
          </SubMenu>
          <SubMenu key="sub3" title="GACTL">
          </SubMenu>
          <SubMenu key="sub4" title="RD">
          </SubMenu>

          <Menu.Divider />
          <Menu.Item key="swagger" title="Swagger"><a href={apiURL}>Swagger API</a></Menu.Item>
          <Menu.Item key="shard_history" title="Shard History"><a href={`${apiURL}/static/shard_history`}>Shard History</a></Menu.Item>
        </Menu>
      </Sider>
        <Content
          style={{
            padding: 24,
            margin: "24px 0px 0 0",
            minHeight: 280,
            overflow: "initial",
          }}>
          <Switch>
            <Route exact path="/">dashboard</Route>
            <Route exact path="/status"><StatusList /></Route>
		  </Switch>
        </Content>
      </Layout>
    </>
  )
}

export default App;
