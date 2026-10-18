import audioIcon from '../assets/images/file/audio.svg'
import excelIcon from '../assets/images/file/excel.svg'
import htmlIcon from '../assets/images/file/html.svg'
import imagesIcon from '../assets/images/file/images.svg'
import pdfIcon from '../assets/images/file/pdf.svg'
import pptIcon from '../assets/images/file/ppt.svg'
import textIcon from '../assets/images/file/text.svg'
import unknownIcon from '../assets/images/file/unknown.svg'
import videoIcon from '../assets/images/file/video.svg'
import wordIcon from '../assets/images/file/word.svg'
import zipIcon from '../assets/images/file/zip.svg'

const extMap = {
  mp3: audioIcon, wav: audioIcon, flac: audioIcon, aac: audioIcon, ogg: audioIcon, wma: audioIcon,
  xls: excelIcon, xlsx: excelIcon, csv: excelIcon,
  html: htmlIcon, htm: htmlIcon, css: htmlIcon,
  jpg: imagesIcon, jpeg: imagesIcon, png: imagesIcon, gif: imagesIcon, bmp: imagesIcon, webp: imagesIcon, svg: imagesIcon, ico: imagesIcon, tiff: imagesIcon,
  pdf: pdfIcon,
  ppt: pptIcon, pptx: pptIcon,
  txt: textIcon, md: textIcon, log: textIcon, json: textIcon, xml: textIcon, yaml: textIcon, yml: textIcon,
  doc: wordIcon, docx: wordIcon,
  mp4: videoIcon, avi: videoIcon, mkv: videoIcon, mov: videoIcon, wmv: videoIcon, flv: videoIcon, webm: videoIcon,
  zip: zipIcon, rar: zipIcon, '7z': zipIcon, tar: zipIcon, gz: zipIcon, bz2: zipIcon,
}

export function getFileIcon(fileName) {
  if (!fileName) return unknownIcon
  const ext = fileName.split('.').pop()?.toLowerCase()
  return extMap[ext] || unknownIcon
}

export function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}
