# retail_stock_counting
Count bottles on a shelf and alert via Line application when the items getting out of stock.

# Requirements
- Tensorflow Object Detection API
- Line Messaging API SDK for Go 
- AWS cli
- Web hosting (use S3 and Cloudfront for testing purpose)

# Installation
First, you have to configure AWS Cli as following:
- https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration

Then create all required stack on AWS include S3, Api Gateway, Lambda Function and CodePipeline by running the following command:
- aws cloudformation create-stack --stack-name StockCountStack --template-body file://cloudformation.yaml

Setup line developer account by following the link below, 
- https://developers.line.biz/en/docs/messaging-api/getting-started/#creating-a-channel

Install Tensorflow Object Detection API by following the line below, 
- https://github.com/tensorflow/models/tree/master/research/object_detection

Copy main.py to models/research/object_detection directory.

Now you can detect bottle and get alert via Line application.

![](images/11628688176542.jpg)
![](images/line.png)
