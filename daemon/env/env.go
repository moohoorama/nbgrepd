package env

import (
	"github.com/caarlos0/env"
)

type Conf struct {
	Bind string `env:"BIND" envDefault:":15050"`

	RoutePrefix   string `env:"ROUTE_PREFIX" envDefault:""`
	BasicAuthName string `env:"BASICAUTH_NAME" envDefault:""`
	BasicAuthPass string `env:"BASICAUTH_PASS" envDefault:""`

	/***************** NB Store 관련 설정 ****************/
	/* bloomfilter 저장 위치*/
	DBPath string `env:"DB_PATH"     envDefault:"./nb.db"`

	/* nbindex등록 안된 크기가 아래보다 커지면, filter에 설정 */
	FilteringThreshold int64 `env:"FILTERING_THRESHOLD" envDefault:"16777216"`
	/* 아래 시간 이상 갱신이 없으면, 그냥 Filtering 해버림 */
	FilteringDurationSec int64 `env:"FILTERING_DURATION_SEC" envDefault:"300"`

	/* 몇글자씩 묶어서 digest 할지 */
	Ngram int `env:"NGRAM" envDefault:"8"`
	/* ngram을 얼마나 촘촘히 할지 */
	Skip int `env:"SKIP" envDefault:"4"`
	/* bloom filter에서 hash 몇번 할까*/
	Numhash int `env:"NUMHASH" envDefault:"7"`
	/* 최대 cardinality 값 */
	MaxCardinality int `env:"MAX_CARDINALITY" envDefault:"262144"`
	/* bloom filter의 크기 */
	NumBytes int `env:"NUMBYTES" envDefault:"524288"`

	TargetPath string `env:"TARGET_PATH" envDefault:"/home/deploy/log/*.log,/usr/local/mkg_api/log/*.log,/usr/local/mkg_httpd/logs/*.log"`

	/* background processing 주기 */
	BGProcSec int `env:"BGPROC_SEC" envDefault:"10"`

	/*  grep을 총괄할 nbgrepd의 주소 */
	MasterAddr         string `env:"MASTER_ADDR" envDefault:"http://tscoke-dev.ay1.krane.9rum.cc:15050"`
	MasterRegistGapSec int64  `env:"MASTER_REGIST_GAP_SEC" envDefault:"60"`
	/* master 등록시 사용할 cluster 이름 */
	ClusterName string `env:"CLUSTER_NAME" envDefault:"TEST_CLUSTER"`

	/* master로써 동작할때, child의 만료 시간
	 * BGProcSec보다는 반드시 길어야 함*/
	ChildExpireSec int64 `env:"CHILD_EXPIRE_SEC" envDefault:"300"`

	HttpTimeoutSec int `env:"HTTP_TIMEOUT_SEC" envDefault:"60"`

	DomainNamePort string `env:"DOMAIN_NAME_PORT"`

	/* keyword가 너무 무난한거라,
	 * 탐색된 chunk가 많으면, grep 연산이 너무 비싸지니
	 * 아래 개수 이상의 filtered chunk가 나오면 grep 중단 */
	MaxFilteredChunkCount int `env:"MAX_FILTERED_CHUNK_COUNT" envDefault:"30"`

	/* tail 작업시, 1800초(30분)내에 수정된 파일만 대상으로 삼음 */
	TailModifyGapSec int64 `env:"TAIL_MODIFY_GAP" envDefault:"1800"`

	/* 아래 크기 만큼만 tail함 */
	TailGapSize int64 `env:"TAIL_GAP_SIZE" envDefault:"8192"`

	BufferPoolSize int `env:"BUFFER_POOL_SIZE" envDefault:"32"`
}

func ParseConfFromEnv() (*Conf, error) {
	conf := &Conf{}
	err := env.Parse(conf)
	if err != nil {
		return nil, err
	}

	if len(conf.DomainNamePort) < 4 {
		hostName, err := GetDomainName()
		if hostName == "" && err != nil {
			return nil, err
		}
		conf.DomainNamePort = hostName + conf.Bind

	}
	return conf, nil
}
