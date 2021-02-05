import React, { useState } from 'react'
// import { useParams } from 'react-router-dom'
// import SongCard from '../Components/SongCard'
// import Loader from '../Components/Loader'
import AxiosGet from '../Components/HttpRequest'

function SearchSongs() {
    const [bool, setBool] = useState(null);
    const [search, setSearch] = useState('');
    const [words, setWords] = useState('');
    
    const  url = 'http://localhost:9000/search/lyrics/'

    let songs = AxiosGet(url, search, words, bool)

    let content = null
   
    if (songs.data){
        content = 
        songs.data.songs.map((song) => {
            <div key={song.id}>
                { song.title }
                { song.artist }
                { song.lyrics.lyrics }
            </div>

            console.log(song.title, song.artist)
        })
        
    }   

    return (
        <div>
            <h2>Search for a song...</h2>
            <form>
                <label>Search: </label>
                    <input className="border rounded-lg" onChange={e => {setSearch(e.target.value)}}></input>
                <label>Words: </label>
                    <input className="border rounded-lg" onChange={e => {setWords(e.target.value)}}></input>
                <button onClick={() => setBool(!bool)}  type="button" className="w-20 border rounded-lg">Submit</button>
            </form>

            { content }
        </div>
    )
}

export default SearchSongs