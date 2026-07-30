"use client";

import React from "react";
import Editor from "@monaco-editor/react";

interface PluginEditorProps {
	value: string;
	onChange: (val: string | undefined) => void;
}

export default function PluginJsonEditor({ value, onChange }: PluginEditorProps) {
	return (
		<div className="border border-gray-800 rounded-xl overflow-hidden glass-panel">
			<div className="bg-gray-900 px-4 py-2 text-xs font-mono text-gray-400 border-b border-gray-800 flex justify-between items-center">
				<span>JSON Configuration Payload</span>
				<span className="text-purple-400">Monaco Editor</span>
			</div>
			<Editor
				height="220px"
				defaultLanguage="json"
				theme="vs-dark"
				value={value}
				onChange={onChange}
				options={{
					minimap: { enabled: false },
					fontSize: 13,
					scrollBeyondLastLine: false,
					automaticLayout: true,
					padding: { top: 12, bottom: 12 },
				}}
			/>
		</div>
	);
}
