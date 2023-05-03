import React  from 'react';
import { Table, Button, message } from "antd";
import axios from "axios";
import { useEffect, useState } from "react";
import { Descriptions, Space } from "antd";

const columns = [
    { title: "Name", dataIndex: "name", render: v => v.toLocaleString()},
    { title: "ServerList", dataIndex: "servers", render: v => v.toLocaleString()},
    { title: "Command", key: "action", render: (v, record) => (<Space size="middle">
                {record.name}
            </Space>)},
];

function StatusList() {
    const [infos, setInfos] = useState([]);
    const [loading, setLoading] = useState(true);

    const loadInfos = () => {
        setLoading(true);
        axios.get("/status")
            .then(
                res => {
                    setLoading(false);
                    setInfos(res.data);
                },
                e => {
                    setLoading(false);
                    message.error(e.toString());
                }
            )
    }

    useEffect(() => {
        loadInfos();
    }, []);


    return (
        <>
            <Descriptions bordered title="Cluster List"
                extra={
                    <Space>
                        <Button onClick={() => loadInfos() }>새로고침</Button>
                    </Space>
                }
            />
            <Table rowKey={"Name"} loading={loading} dataSource={infos} columns={columns} pagination={50} />
        </>
    )
}

export { StatusList };
