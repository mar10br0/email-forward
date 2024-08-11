call ./secrets.cmd
set AWS_ACCESS_KEY=%TOKENGEN_AWS_ACCESS_KEY%
set AWS_SECRET_KEY=%TOKENGEN_AWS_SECRET_KEY%
set LAMBDA_BUCKET=email-forward-lambda-package
token-gen.exe