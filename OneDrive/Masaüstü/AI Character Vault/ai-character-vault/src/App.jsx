import { useState } from 'react';
import { useNotes } from './hooks/useNotes';
import NoteForm from './components/NoteForm';
import NoteCard from './components/NoteCard';

function App() {
  const { notes, loading, addNote, updateNote, deleteNote } = useNotes();
  const [editingNote, setEditingNote] = useState(null);
  const [filter, setFilter] = useState('all');
  const [searchTerm, setSearchTerm] = useState('');

  const filteredNotes = notes.filter((note) => {
    const matchesFilter = filter === 'all' || note.category === filter || note.priority === filter;
    const matchesSearch = note.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         note.content.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  const handleAddNote = (noteData) => {
    addNote(noteData);
  };

  const handleUpdateNote = (noteData) => {
    updateNote(editingNote.id, noteData);
    setEditingNote(null);
  };

  const handleEdit = (note) => {
    setEditingNote(note);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleCancelEdit = () => {
    setEditingNote(null);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-background-50 flex items-center justify-center">
        <p className="text-background-600">Yükleniyor...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background-50 py-8 px-4">
      <div className="max-w-6xl mx-auto">
        <header className="text-center mb-8">
          <h1 className="text-4xl font-bold text-background-900 mb-2">
            🤖 Akıllı Not Uygulaması
          </h1>
          <p className="text-background-600">
            Yapay zeka ile otomatik kategori ve öncelik belirleme
          </p>
        </header>

        <NoteForm
          onSubmit={editingNote ? handleUpdateNote : handleAddNote}
          editingNote={editingNote}
          onCancel={handleCancelEdit}
        />

        <div className="bg-white rounded-2xl shadow-soft p-6 mb-6">
          <div className="flex flex-col md:flex-row gap-4 mb-4">
            <div className="flex-1">
              <input
                type="text"
                placeholder="Notlarda ara..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full px-4 py-3 border border-background-200 rounded-xl focus:ring-2 focus:ring-primary-400 focus:border-transparent outline-none bg-background-50"
              />
            </div>
            <div className="flex gap-2 flex-wrap">
              <button
                onClick={() => setFilter('all')}
                className={`px-4 py-2 rounded-xl transition-all ${
                  filter === 'all' 
                    ? 'bg-primary-500 text-white shadow-soft' 
                    : 'bg-background-100 text-background-700 hover:bg-background-200'
                }`}
              >
                Tümü
              </button>
              <button
                onClick={() => setFilter('iş')}
                className={`px-4 py-2 rounded-xl transition-all ${
                  filter === 'iş' 
                    ? 'bg-primary-500 text-white shadow-soft' 
                    : 'bg-background-100 text-background-700 hover:bg-background-200'
                }`}
              >
                İş
              </button>
              <button
                onClick={() => setFilter('kişisel')}
                className={`px-4 py-2 rounded-xl transition-all ${
                  filter === 'kişisel' 
                    ? 'bg-primary-500 text-white shadow-soft' 
                    : 'bg-background-100 text-background-700 hover:bg-background-200'
                }`}
              >
                Kişisel
              </button>
              <button
                onClick={() => setFilter('yüksek')}
                className={`px-4 py-2 rounded-xl transition-all ${
                  filter === 'yüksek' 
                    ? 'bg-secondary-500 text-white shadow-soft' 
                    : 'bg-background-100 text-background-700 hover:bg-background-200'
                }`}
              >
                Yüksek Öncelik
              </button>
            </div>
          </div>
          <p className="text-sm text-background-500">
            Toplam {filteredNotes.length} not
          </p>
        </div>

        {filteredNotes.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-background-500 text-lg">
              {searchTerm || filter !== 'all' ? 'Filtrelere uygun not bulunamadı.' : 'Henüz not eklenmedi. İlk notunuzu ekleyin!'}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
           {filteredNotes.map((note) => (
              <NoteCard
                key={note.id}
                note={note}
                onEdit={handleEdit}
                onDelete={deleteNote}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
