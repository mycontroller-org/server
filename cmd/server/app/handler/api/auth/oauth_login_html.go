package auth

// https://unpkg.com/@patternfly/patternfly/patternfly.css => https://unpkg.com/@patternfly/patternfly@4.196.7/patternfly.css
const OAuthLoginPageHTML = `
<!DOCTYPE html>
<html lang="en" class="pf-m-redhat-font">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="stylesheet" href="fonts.css" />
    <!-- Include latest PatternFly CSS via CDN -->
    <link
      rel="stylesheet"
      href="https://unpkg.com/@patternfly/patternfly@4.196.7/patternfly.css"
      crossorigin="anonymous"
    />
    <link rel="stylesheet" href="style.css" />
    <title>MyController Server oAuth Login Page</title>
  </head>
  <body>
    <div class="pf-c-background-image"></div>
    <div class="pf-c-login">
      <div class="pf-c-login__container">
        <main class="pf-c-login__main">
          <header class="pf-c-login__main-header">
            <h1 class="pf-c-title pf-m-3xl">Welcome to MyController</h1>
            <p class="pf-c-login__main-header-desc">
              Login to your account to complete the oAuth request
            </p>
          </header>
          <div class="pf-c-login__main-body">
            <form novalidate class="pf-c-form" method="post">
              <p class="pf-c-form__helper-text pf-m-error pf-m-hidden">
                <i
                  class="fas fa-exclamation-circle pf-c-form__helper-text-icon"
                  aria-hidden="true"
                ></i>
                Invalid login credentials.
              </p>
              <div class="pf-c-form__group">
                <label class="pf-c-form__label" for="login-demo-form-username">
                  <span class="pf-c-form__label-text">Username</span>
                  <span class="pf-c-form__label-required" aria-hidden="true"
                    >&#42;</span
                  >
                </label>
                <input
                  class="pf-c-form-control"
                  required
                  input="true"
                  type="text"
                  id="username"
                  name="username"
                />
              </div>
              <div class="pf-c-form__group">
                <label class="pf-c-form__label" for="login-demo-form-password">
                  <span class="pf-c-form__label-text">Password</span>
                  <span class="pf-c-form__label-required" aria-hidden="true"
                    >&#42;</span
                  >
                </label>
                <input
                  class="pf-c-form-control"
                  required
                  input="true"
                  type="password"
                  id="password"
                  name="password"
                />
              </div>
              <div class="pf-c-form__group pf-m-action">
                <button
                  class="pf-c-button pf-m-primary pf-m-block"
                  type="submit"
                >
                  Log In
                </button>
              </div>
            </form>
          </div>
        </main>
      </div>
    </div>
  </body>
</html>
`
