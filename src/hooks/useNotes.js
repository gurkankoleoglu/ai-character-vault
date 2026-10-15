import { useState, useEffect } from 'react';
import { getNotes, addNote, updateNote, deleteNote } from '../utils/localStorage';

export const useNotes = () => {
  const [notes, setNotes] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadedNotes = getNotes();
    setNotes(loadedNotes);
    setLoading(false);
  }, []);

  const handleAddNote = (note) => {
    const newNote = addNote(note);
    setNotes([...notes, newNote]);
    return newNote;
  };

  const handleUpdateNote = (id, updatedNote) => {
    const updated = updateNote(id, updatedNote);
    if (updated) {
      setNotes(notes.map(note => note.id === id ? updated : note));
    }
    return updated;
  };

  const handleDeleteNote = (id) => {
    const filtered = deleteNote(id);
    setNotes(filtered);
  };

  return {
    notes,
    loading,
    addNote: handleAddNote,
    updateNote: handleUpdateNote,
    deleteNote: handleDeleteNote,
  };
};
