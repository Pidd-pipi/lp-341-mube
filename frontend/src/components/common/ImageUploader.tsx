import { Upload, Button } from 'antd';
import { UploadOutlined } from '@ant-design/icons';

interface Props {
  value?: string;
  onChange?: (url: string) => void;
}

// 影像图片上传（结果录入）
export default function ImageUploader({ value, onChange }: Props) {
  return (
    <Upload
      accept="image/*"
      maxCount={1}
      listType="picture"
      beforeUpload={() => false}
      onChange={(info) => {
        const file = info.fileList[0]?.originFileObj;
        if (file && onChange) {
          const reader = new FileReader();
          reader.onload = () => onChange(String(reader.result));
          reader.readAsDataURL(file as Blob);
        }
      }}
    >
      <Button icon={<UploadOutlined />}>上传影像</Button>
      {value && <span style={{ marginLeft: 8, color: '#999' }}>已选择图片</span>}
    </Upload>
  );
}
