package main

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Admin Login</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    background: #111;
    color: #e0e0e0;
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .login-card {
    background: #1a1a1a;
    border: 1px solid #2a2a2a;
    border-radius: 16px;
    padding: 40px;
    width: 100%;
    max-width: 400px;
    box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  }
  .login-title {
    font-size: 22px;
    font-weight: 600;
    color: #c8a97e;
    text-align: center;
    margin-bottom: 8px;
    letter-spacing: 1px;
  }
  .login-subtitle {
    font-size: 13px;
    color: #666;
    text-align: center;
    margin-bottom: 28px;
  }
  .field {
    margin-bottom: 18px;
  }
  .field label {
    display: block;
    font-size: 13px;
    color: #999;
    margin-bottom: 6px;
    font-weight: 500;
  }
  .field input {
    width: 100%;
    padding: 12px 14px;
    background: #111;
    border: 1px solid #333;
    border-radius: 8px;
    color: #e0e0e0;
    font-size: 15px;
    font-family: inherit;
    transition: border-color 0.2s;
  }
  .field input:focus { outline: none; border-color: #c8a97e; }
  .btn {
    width: 100%;
    padding: 12px;
    background: #c8a97e;
    color: #111;
    border: none;
    border-radius: 8px;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
    font-family: inherit;
    margin-top: 8px;
  }
  .btn:hover { background: #d4b88f; }
  .error {
    background: rgba(229, 115, 115, 0.15);
    border: 1px solid rgba(229, 115, 115, 0.3);
    color: #e57373;
    padding: 10px 16px;
    border-radius: 8px;
    font-size: 13px;
    margin-bottom: 20px;
    text-align: center;
  }
</style>
</head>
<body>
<div class="login-card">
  <div class="login-title">Gallery Admin</div>
  <div class="login-subtitle">Sign in to manage your galleries</div>
  {{if .Error}}
  <div class="error">{{.Error}}</div>
  {{end}}
  <form method="POST" action="/admin/login">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
    <div class="field">
      <label>Username</label>
      <input type="text" name="username" autocomplete="username" autofocus required>
    </div>
    <div class="field">
      <label>Password</label>
      <input type="password" name="password" autocomplete="current-password" required>
    </div>
    <button type="submit" class="btn">Sign In</button>
  </form>
</div>
</body>
</html>`
