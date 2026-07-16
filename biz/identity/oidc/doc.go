// Package oidc bridges a standard OpenID Connect authorization-code flow into
// the post-auth identity mapping package.
//
// It handles provider metadata, authorization URLs, code exchange, ID-token
// verification, nonce checking, optional one-time login state storage, and
// assertion construction. It does not issue application sessions, cookies, or
// account-management screens.
//
// 包 oidc 将标准 OpenID Connect 授权码流程衔接到认证后身份映射包。
//
// 它处理提供方元数据、授权 URL、代码交换、ID 令牌验证、nonce 校验、可选的一次性登录
// 状态存储和断言构造；但不签发应用会话、Cookie 或账户管理界面。
package oidc
