import { Button, Input, Form, Space, Col, Row, message } from "antd";
import { useHistory } from 'react-router';
import axios from 'axios';
import { useEffect, useState } from "react";

function Login() {
    const history = useHistory();
    const [ssoURL, setSSOURL] = useState("");
    const [loggingIn, setLoggingIn] = useState(false);

    useEffect(() => {
        axios.get("/auth/sso")
            .then(res => {
                setSSOURL(res.data.url);
            })
            .catch(e => {
                message.error(e.toString());
                console.log(e);
            })
    }, [])

    const onFinish = (values) => {
        let pass = values.XAdminPass ? values.XAdminPass : "";

        const params = new URLSearchParams();
        params.append('x-admin-pass', pass);

        setLoggingIn(true);
        axios.post("/auth/adminpass/login", params)
            .then(() => {
                setLoggingIn(false);
                history.go(0);
            })
            .catch(e => {
                message.error(e.toString());
                console.log(e);
                setLoggingIn(false);
            })
    };
    
    const onFinishFailed = e => {
        message.error(e.toString());
        console.log(e);
    };

    return (
        <Space direction="vertical" style={{height: '100%', justifyContent: "center"}}>
            <Space direction="horizontal" style={{ width: '100%', justifyContent: 'center'}}>
                <Row align="middle" justify="center" style={{ minWidth: "700px" }} >
                <Col span={11} style={{ paddingRight: "5px", borderRight: "1px solid #ccc" }}>
                <Form
                    title="X-Admin-Pass"
                    name="basic"
                    labelCol={{ span: 9 }}
                    wrapperCol={{ span: 12 }}
                    onFinish={onFinish}
                    onFinishFailed={onFinishFailed}
                    autoComplete="off"
                >
                    <Form.Item label="X-Admin-Pass" name="XAdminPass">
                        <Input />
                    </Form.Item>
                    
                    <Form.Item wrapperCol={{ offset: 9, span: 5 }}>
                        <Button type="primary" htmlType="submit" disabled={loggingIn}> Login </Button>
                    </Form.Item>
                </Form>
                </Col>
                <Col span={11}>
                    <Space direction="horizontal" style={{ width: "100%", justifyContent: 'center' }}>
                        <Button onClick={() => setLoggingIn(true)} loading={loggingIn} href={ssoURL} type="primary"> Login via 9rum SSO </Button>
                    </Space>
                </Col>
                </Row>
        </Space>
      </Space>
    )
}

export { Login };

