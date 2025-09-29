/**
 * Jest 测试设置文件
 */

// 设置测试环境变量
process.env['NODE_ENV'] = 'test';

// 模拟 console.log 以避免测试输出干扰
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
};

// 设置测试超时时间
jest.setTimeout(10000);
