# retail_stock_counting
Count bottles on a shelf and alert via Line application when the items getting out of stock.

# Requirements
- Tensorflow Object Detection API
- Line Messaging API SDK for Go 
- AWS cli

# Installation
You can create all required stack on AWS include S3, Api Gateway and Lambda Function by running following command:
- aws cloudformation create-stack --stack-name StockCountStack --template-body file://sampletemplate.json 
