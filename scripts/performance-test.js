import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// 自定义指标
export let errorRate = new Rate('errors');
export let loginDuration = new Trend('login_duration');
export let apiDuration = new Trend('api_duration');

// 测试配置
export let options = {
  stages: [
    { duration: '2m', target: 10 },   // 预热阶段
    { duration: '5m', target: 50 },   // 负载增加
    { duration: '10m', target: 100 }, // 稳定负载
    { duration: '5m', target: 200 },  // 峰值负载
    { duration: '3m', target: 0 },    // 负载下降
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% 请求响应时间 < 500ms
    http_req_failed: ['rate<0.01'],   // 错误率 < 1%
    errors: ['rate<0.05'],            // 自定义错误率 < 5%
  },
};

const BASE_URL = 'http://localhost:8080';

// 测试数据
const users = [
  { username: 'testuser1', password: 'password123' },
  { username: 'testuser2', password: 'password123' },
  { username: 'testuser3', password: 'password123' },
];

let authToken = '';

export function setup() {
  // 测试前的准备工作
  console.log('🚀 开始性能测试准备');
  
  // 创建测试用户
  const testUser = {
    username: 'perftest' + Date.now(),
    email: 'perftest@example.com',
    password: 'password123'
  };
  
  const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(registerRes, {
    '用户注册成功': (r) => r.status === 201,
  });
  
  return { testUser };
}

export default function(data) {
  // 1. 用户登录测试
  testLogin(data.testUser);
  
  // 2. API 端点测试
  if (authToken) {
    testUserAPIs();
    testHealthCheck();
  }
  
  sleep(1);
}

function testLogin(user) {
  const loginPayload = {
    username: user.username,
    password: user.password
  };
  
  const loginStart = Date.now();
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify(loginPayload), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const loginSuccess = check(loginRes, {
    '登录状态码 200': (r) => r.status === 200,
    '返回 JWT Token': (r) => r.json('data.token') !== '',
  });
  
  if (loginSuccess) {
    authToken = loginRes.json('data.token');
    loginDuration.add(Date.now() - loginStart);
  } else {
    errorRate.add(1);
  }
}

function testUserAPIs() {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
  };
  
  // 获取用户列表
  const getUsersStart = Date.now();
  const getUsersRes = http.get(`${BASE_URL}/api/v1/users`, { headers });
  
  check(getUsersRes, {
    '获取用户列表成功': (r) => r.status === 200,
    '返回用户数据': (r) => Array.isArray(r.json('data')),
  }) || errorRate.add(1);
  
  apiDuration.add(Date.now() - getUsersStart);
  
  // 创建新用户
  const newUser = {
    username: 'apitest' + Math.random().toString(36).substr(2, 9),
    email: 'apitest' + Math.random().toString(36).substr(2, 9) + '@example.com',
    password: 'password123'
  };
  
  const createUserStart = Date.now();
  const createUserRes = http.post(`${BASE_URL}/api/v1/users`, JSON.stringify(newUser), { headers });
  
  check(createUserRes, {
    '创建用户成功': (r) => r.status === 201,
    '返回新用户信息': (r) => r.json('data.id') > 0,
  }) || errorRate.add(1);
  
  apiDuration.add(Date.now() - createUserStart);
  
  if (createUserRes.status === 201) {
    const userId = createUserRes.json('data.id');
    
    // 获取单个用户
    const getUserStart = Date.now();
    const getUserRes = http.get(`${BASE_URL}/api/v1/users/${userId}`, { headers });
    
    check(getUserRes, {
      '获取单个用户成功': (r) => r.status === 200,
      '用户 ID 匹配': (r) => r.json('data.id') === userId,
    }) || errorRate.add(1);
    
    apiDuration.add(Date.now() - getUserStart);
    
    // 更新用户
    const updateUser = {
      username: newUser.username + '_updated',
      email: newUser.email,
    };
    
    const updateUserStart = Date.now();
    const updateUserRes = http.put(`${BASE_URL}/api/v1/users/${userId}`, JSON.stringify(updateUser), { headers });
    
    check(updateUserRes, {
      '更新用户成功': (r) => r.status === 200,
    }) || errorRate.add(1);
    
    apiDuration.add(Date.now() - updateUserStart);
    
    // 删除用户
    const deleteUserStart = Date.now();
    const deleteUserRes = http.del(`${BASE_URL}/api/v1/users/${userId}`, null, { headers });
    
    check(deleteUserRes, {
      '删除用户成功': (r) => r.status === 200,
    }) || errorRate.add(1);
    
    apiDuration.add(Date.now() - deleteUserStart);
  }
}

function testHealthCheck() {
  const healthRes = http.get(`${BASE_URL}/health`);
  
  check(healthRes, {
    '健康检查成功': (r) => r.status === 200,
    '所有服务健康': (r) => {
      const health = r.json();
      return health.mysql === 'healthy' && 
             health.redis === 'healthy' && 
             health.rabbitmq === 'healthy';
    },
  }) || errorRate.add(1);
}

export function teardown(data) {
  console.log('🏁 性能测试清理完成');
}

// 使用命令运行测试:
// k6 run --out influxdb=http://localhost:8086/mydb scripts/performance-test.js 