import { useState } from 'react';
import { analyzeNote } from '../utils/aiAnalyzer';

export default function NoteForm({ onSubmit, editingNote, onCancel }) {
  const [title, setTitle] = useState(editingNote?.title || '');
  const [content, setContent] = useState(editingNote?.content || '');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!title.trim() || !content.trim()) return;

    const aiAnalysis = analyzeNote(content);
    
    onSubmit({
      title,
      content,
      ...aiAnalysis,
    });

    if (!editingNote) {
      setTitle('');
      setContent('');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="bg-white rounded-2xl shadow-soft p-6 mb-6">
      <h2 className="text-2xl font-bold mb-4 text-background-900">
        {editingNote ? 'Notu Düzenle' : 'Yeni Not Ekle'}
      </h2>
      <div className="mb-4">
        <label className="block text-sm font-medium text-background-700 mb-2">
          Başlık
        </label>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full px-4 py-3 border border-background-200 rounded-xl focus:ring-2 focus:ring-primary-400 focus:border-transparent outline-none bg-background-50"
          placeholder="Not başlığı..."
          required
        />
      </div>
      <div className="mb-4">
        <label className="block text-sm font-medium text-background-700 mb-2">
          İçerik
        </label>
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          className="w-full px-4 py-3 border border-background-200 rounded-xl focus:ring-2 focus:ring-primary-400 focus:border-transparent outline-none h-32 resize-none bg-background-50"
          placeholder="Not içeriğini yazın... AI otomatik kategori ve öncelik belirleyecek."
          required
        />
      </div>
      <div className="flex gap-3">
        <button
          type="submit"
          className="bg-primary-500 hover:bg-primary-600 text-white font-medium py-2 px-6 rounded-xl transition-all shadow-soft"
        >
          {editingNote ? 'Güncelle' : 'Ekle'}
        </button>
        {editingNote && (
          <button
            type="button"
            onClick={onCancel}
            className="bg-background-200 hover:bg-background-300 text-background-800 font-medium py-2 px-6 rounded-xl transition-all"
          >
            İptal
          </button>
        )}
      </div>
    </form>
  );
}
