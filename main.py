import base64
import cv2
import json
import os
import pathlib
import requests
import numpy as np
import tensorflow as tf

from PIL import Image
from io import BytesIO
from IPython.display import display

from object_detection.utils import ops as utils_ops
from object_detection.utils import label_map_util
from object_detection.utils import visualization_utils as vis_util

# patch tf1 into `utils.ops`
utils_ops.tf = tf.compat.v1

# Patch the location of gfile
tf.gfile = tf.io.gfile

if "models" in pathlib.Path.cwd().parts:
    while "models" in pathlib.Path.cwd().parts:
        os.chdir('..')

VIDEO_NAME = 'bp_video.mp4'
CWD_PATH = os.getcwd()
PATH_TO_VIDEO = os.path.join(CWD_PATH,VIDEO_NAME)

PATH_TO_LABELS = 'models/research/object_detection/data/mscoco_label_map.pbtxt'
category_index = label_map_util.create_category_index_from_labelmap(PATH_TO_LABELS, use_display_name=True)

model_name = 'faster_rcnn_inception_v2_coco_2018_01_28'

def load_model():
    base_url = 'http://download.tensorflow.org/models/object_detection/'
    model_file = model_name + '.tar.gz'
    model_dir = tf.keras.utils.get_file(
        fname=model_name, 
        origin=base_url + model_file,
        untar=True)

    model_dir = pathlib.Path(model_dir)/"saved_model"

    model = tf.saved_model.load(str(model_dir))
    model = model.signatures['serving_default']

    return model

def run_inference_for_single_image(model, image):
    image = np.asarray(image)
    # The input needs to be a tensor, convert it using `tf.convert_to_tensor`.
    input_tensor = tf.convert_to_tensor(image)
    # The model expects a batch of images, so add an axis with `tf.newaxis`.
    input_tensor = input_tensor[tf.newaxis,...]

    # Run inference
    output_dict = model(input_tensor)

    # All outputs are batches tensors.
    # Convert to numpy arrays, and take index [0] to remove the batch dimension.
    # We're only interested in the first num_detections.
    num_detections = int(output_dict.pop('num_detections'))
    output_dict = {key:value[0, :num_detections].numpy() 
                    for key,value in output_dict.items()}
    output_dict['num_detections'] = num_detections

    # detection_classes should be ints.
    output_dict['detection_classes'] = output_dict['detection_classes'].astype(np.int64)

    # Handle models with masks:
    if 'detection_masks' in output_dict:
    # Reframe the the bbox mask to the image size.
        detection_masks_reframed = utils_ops.reframe_box_masks_to_image_masks(
            output_dict['detection_masks'], output_dict['detection_boxes'],
            image.shape[0], image.shape[1])
        detection_masks_reframed = tf.cast(detection_masks_reframed > 0.5,tf.uint8)
        output_dict['detection_masks_reframed'] = detection_masks_reframed.numpy()

    return output_dict

def show_inference(model, image_np):
    # the array based representation of the image will be used later in order to prepare the
    # result image with boxes and labels on it.
    #  image_np = np.array(Image.open(image_path))
    # Actual detection.
    # image_np = cv2.resize(image_np, (800, 600))
    output_dict = run_inference_for_single_image(model, image_np)
    # Visualization of the results of a detection.
    vis_util.visualize_boxes_and_labels_on_image_array(
        image_np,
        output_dict['detection_boxes'],
        output_dict['detection_classes'],
        output_dict['detection_scores'],
        category_index,
        instance_masks=output_dict.get('detection_masks_reframed', None),
        use_normalized_coordinates=True,
        line_thickness=8)

    final_score = np.squeeze(output_dict['detection_scores'])
    scores = final_score[final_score>0.5]
    classes = output_dict['detection_classes'][:len(scores)]

    #print(output_dict['detection_classes'].shape)

    return image_np, classes
def send_msg(msg):
    API_ENDPOINT = "https://ozof4y1ld6.execute-api.ap-southeast-1.amazonaws.com/test/notify/"
    response = requests.get(API_ENDPOINT+msg)
    print(msg, ': ', response.status_code)

def send_img(image_np):
    API_ENDPOINT = 'https://ozof4y1ld6.execute-api.ap-southeast-1.amazonaws.com/test/upload'
    buffered = BytesIO()

    image_np = cv2.cvtColor(image_np, cv2.COLOR_BGR2RGB)
    img = Image.fromarray(image_np, 'RGB')
    img.save(buffered, format="JPEG")
    img_str = base64.b64encode(buffered.getvalue())

    data = {'imageBase64': img_str.decode("utf-8")}

    r = requests.post(url=API_ENDPOINT, data=json.dumps(data))

    print(json.loads(r.text))

def count_bottle(classes):
    return np.count_nonzero(classes == 44)

def main():
    th = 4
    count = 0
    notified = False
    bottle_count = th

    fourcc = cv2.VideoWriter_fourcc('M','J','P','G')
    out = cv2.VideoWriter('output.avi', fourcc, 10.0, (800, 600), True)
    cap = cv2.VideoCapture(PATH_TO_VIDEO)

    only_bottle = lambda x: len(np.unique(x)) == 1 and np.unique(x)[0] == 44
    detection_model = load_model()
    
    while(cap.isOpened()):
        count += 1
        
        ret, frame = cap.read()

        if ret == True:
            image_np = cv2.resize(frame, (800, 600))
            if not count % 10:

                count = 0
                image_np, classes = show_inference(detection_model, image_np)

                if only_bottle(classes):
                    new_bottle_count = count_bottle(classes)
                    bottle_count = (bottle_count + new_bottle_count) * 0.6
                    print("class: ", classes)

                    print('bottle_count/score: ', new_bottle_count, '|', bottle_count)

                    if (bottle_count < th) and not notified:
                        notified = True
                        send_msg("items are running out")
                        send_img(image_np)
                        
                    elif (bottle_count >= th) and notified:
                        notified = False
                        send_msg('items are restored')
                        send_img(image_np)
                    
                    else: pass

                # save
                out.write(image_np)

                # show 
                cv2.imshow('object detection', image_np)
           
            if cv2.waitKey(1) == ord('q'):
                break
        else:
            break
    cap.release()
    out.release()
   
    cv2.destroyAllWindows()

# exec
if __name__ == "__main__":
    main()


