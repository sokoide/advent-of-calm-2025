import Editor from '@monaco-editor/react';

interface CodeEditorProps {
  value: string;
  language: string;
  onChange: (value: string | undefined) => void;
  readOnly?: boolean;
}

const CodeEditor = ({ value, language, onChange, readOnly = false }: CodeEditorProps) => {
  return (
    <Editor
      height="100%"
      language={language}
      theme="vs-dark"
      value={value}
      onChange={onChange}
      options={{
        minimap: { enabled: false },
        fontSize: 13,
        readOnly,
        automaticLayout: true,
      }}
    />
  );
};

export default CodeEditor;
